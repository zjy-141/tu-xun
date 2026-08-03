package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// generateVerifyCode 生成 12 位随机防伪核销码 [A-Z0-9]
func generateVerifyCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 12
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

type ExchangeSvc struct{}

// Claim 兑换奖品（幂等）
func (e *ExchangeSvc) Claim(params ExchangeCreateParams) (ResponseIS, error) {
	// 基础校验
	if params.Quantity <= 0 {
		return ResponseIS{}, common.ErrNew(errors.New("兑换数量必须为正数"), common.ParamErr)
	}

	// 幂等键检查
	if params.IdempotencyKey != "" {
		var existing model.Exchange
		if err := model.DB.Where("idempotency_key = ? AND user_id = ?", params.IdempotencyKey, params.UserID).First(&existing).Error; err == nil {
			// 同键同内容 -> 返回首次结果
			if existing.GoodID == params.GoodID && existing.Quantity == params.Quantity {
				return ResponseIS{ID: existing.ID, Status: existing.Status}, nil
			}
			// 同键不同内容 -> 409
			return ResponseIS{}, common.ErrNew(errors.New("幂等键冲突：同键不同内容"), common.ConflictErr)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return ResponseIS{}, common.ErrNew(err, common.SysErr)
		}
	}

	tx := model.DB.Begin()
	var err error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. 锁定奖品并查询（加行锁防止并发），必须是上架状态
	var good model.Good
	if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "name", "need_score", "stock").
		Where("id = ? AND status = ?", params.GoodID, "in_store").
		First(&good).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ResponseIS{}, common.ErrNew(errors.New("奖品不存在或已下架"), common.ParamErr)
		}
		return ResponseIS{}, common.ErrNew(err, common.SysErr)
	}

	// 2. 计算消耗积分
	cost := good.NeedScore * params.Quantity

	// 3. 锁定用户并查询余额
	var user model.User
	if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("score_count").
		Where("id = ?", params.UserID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ResponseIS{}, common.ErrNew(errors.New("用户不存在"), common.ParamErr)
		}
		return ResponseIS{}, common.ErrNew(err, common.SysErr)
	}

	// 4. 校验库存和积分
	if good.Stock < params.Quantity {
		return ResponseIS{}, common.ErrNew(errors.New("奖品库存不足"), common.ParamErr)
	}
	if user.ScoreCount < cost {
		return ResponseIS{}, common.ErrNew(errors.New("用户积分不足"), common.ParamErr)
	}

	// 5. 扣减库存（条件更新，防止并发超卖）
	stockResult := tx.Model(&model.Good{}).
		Where("id = ? AND stock >= ?", params.GoodID, params.Quantity).
		Update("stock", gorm.Expr("stock - ?", params.Quantity))
	if stockResult.Error != nil {
		return ResponseIS{}, common.ErrNew(stockResult.Error, common.SysErr)
	}
	if stockResult.RowsAffected == 0 {
		return ResponseIS{}, common.ErrNew(errors.New("奖品库存不足，请重试"), common.ParamErr)
	}

	// 6. 扣减积分（条件更新）
	scoreResult := tx.Model(&model.User{}).
		Where("id = ? AND score_count >= ?", params.UserID, cost).
		Update("score_count", gorm.Expr("score_count - ?", cost))
	if scoreResult.Error != nil {
		return ResponseIS{}, common.ErrNew(scoreResult.Error, common.SysErr)
	}
	if scoreResult.RowsAffected == 0 {
		return ResponseIS{}, common.ErrNew(errors.New("用户积分不足，请重试"), common.ParamErr)
	}

	// 7. 创建兑换记录（状态为 pending）
	exchange := &model.Exchange{
		GoodID:         params.GoodID,
		UserID:         params.UserID,
		Quantity:       params.Quantity,
		ScoreCost:      cost,
		Status:         "pending",
		VerifyCode:     generateVerifyCode(),
		IdempotencyKey: params.IdempotencyKey,
	}
	if err = tx.Create(exchange).Error; err != nil {
		return ResponseIS{}, common.ErrNew(err, common.SysErr)
	}

	// 8. 创建积分日志（记录扣减后的余额）
	scoreLog := &model.ScoreLog{
		UserID:      params.UserID,
		Delta:       -cost,
		Balance:     user.ScoreCount - cost,
		Reason:      "exchange",
		RelatedID:   exchange.ID,
		RelatedType: "exchange",
		Remark:      fmt.Sprintf("兑换奖品 %d（%s），数量 %d，消耗积分 %d", params.GoodID, good.Name, params.Quantity, cost),
	}
	if err = tx.Create(scoreLog).Error; err != nil {
		return ResponseIS{}, common.ErrNew(err, common.SysErr)
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return ResponseIS{}, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}

	return ResponseIS{
		ID:         exchange.ID,
		Status:     exchange.Status,
		VerifyCode: exchange.VerifyCode,
	}, nil
}

// List 获取兑奖记录
func (e *ExchangeSvc) List(params ExchangeListParams) (ExchangeItemPage, error) {
	var exchanges []model.Exchange
	var total int64

	query := model.DB.Model(&model.Exchange{}).
		Where("user_id = ?", params.UserID)

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return ExchangeItemPage{}, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("Good").
		Order("id DESC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&exchanges).Error; err != nil {
		return ExchangeItemPage{}, common.ErrNew(err, common.SysErr)
	}

	resp := ExchangeItemPage{
		Total: total,
		List:  make([]ExchangeItem, 0, len(exchanges)),
	}
	for _, ec := range exchanges {
		resp.List = append(resp.List, ExchangeItem{
			ID: ec.ID,
			Good: GoodBrief{
				ID:         ec.Good.ID,
				Name:       ec.Good.Name,
				Image: Media{ThumbURL: ec.Good.ThumbURL, Width: ec.Good.ThumbWidth, Height: ec.Good.ThumbHeight},
				ScorePrice: ec.Good.NeedScore,
			},
			Quantity:   ec.Quantity,
			ScoreCost:  ec.ScoreCost,
			Status:     ec.Status,
				VerifyCode: ec.VerifyCode,
			ExchangeAt: ec.ExchangeAt,
			CreatedAt:  &ec.CreatedAt,
		})
	}

	return resp, nil
}
