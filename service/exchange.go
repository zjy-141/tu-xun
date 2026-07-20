package service

import (
	"errors"
	"fmt"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExchangeSvc struct{}

// Claim 兑换奖品
func (e *ExchangeSvc) Claim(info ExchangeClaim) (resp ResponseIS, err error) {
	// 基础校验
	if info.Quantity <= 0 {
		return resp, common.ErrNew(errors.New("兑换数量必须为正数"), common.ParamErr)
	}

	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		// 关键：只要 err != nil，就回滚事务
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. 锁定奖品并查询（加行锁防止并发）
	var good model.Good
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "name", "need_score", "stock").
		Where("id = ? AND status = ?", info.GoodID, "inStore").
		First(&good).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("奖品不存在"), common.ParamErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 2. 计算消耗积分
	cost := good.NeedScore * info.Quantity

	// 3. 锁定用户并查询余额
	var user model.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("score_count").
		Where("id = ?", info.UserID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("用户不存在"), common.ParamErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 4. 校验库存和积分（乐观锁+条件更新保证原子性，但此处先做业务校验，减少无效更新）
	if good.Stock < info.Quantity {
		return resp, common.ErrNew(errors.New("奖品库存不足"), common.ParamErr)
	}
	if user.ScoreCount < cost {
		return resp, common.ErrNew(errors.New("用户积分不足"), common.ParamErr)
	}

	// 5. 扣减库存（条件更新，防止并发超卖）
	stockResult := tx.Model(&model.Good{}).
		Where("id = ? AND stock >= ?", info.GoodID, info.Quantity).
		Update("stock", gorm.Expr("stock - ?", info.Quantity))
	if stockResult.Error != nil {
		return resp, common.ErrNew(stockResult.Error, common.SysErr)
	}
	if stockResult.RowsAffected == 0 {
		// 理论上不会发生，因为前面已校验，但并发冲突时可能
		return resp, common.ErrNew(errors.New("奖品库存不足，请重试"), common.ParamErr)
	}

	// 6. 扣减积分（条件更新）
	scoreResult := tx.Model(&model.User{}).
		Where("id = ? AND score_count >= ?", info.UserID, cost).
		Update("score_count", gorm.Expr("score_count - ?", cost))
	if scoreResult.Error != nil {
		return resp, common.ErrNew(scoreResult.Error, common.SysErr)
	}
	if scoreResult.RowsAffected == 0 {
		return resp, common.ErrNew(errors.New("用户积分不足，请重试"), common.ParamErr)
	}

	// 7. 创建兑换记录（状态为 pending）
	exchange := &model.Exchange{
		GoodID:    info.GoodID,
		UserID:    info.UserID,
		Quantity:  info.Quantity,
		ScoreCost: cost,
		Status:    "pending",
		// ExchangeAt 留空（或零值），取货时再更新
	}
	if err := tx.Create(exchange).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 8. 创建积分日志（记录扣减后的余额）
	// 因为已查询过 user 的旧余额，扣减后新余额为 user.ScoreCount - cost
	scoreLog := &model.ScoreLog{
		UserID:      info.UserID,
		Delta:       -cost,
		Balance:     user.ScoreCount - cost, // 直接计算，无需二次查询
		Reason:      "exchange",
		RelatedID:   exchange.ID,
		RelatedType: "exchange",
		Remark:      fmt.Sprintf("兑换奖品 %d（%s），数量 %d，消耗积分 %d", info.GoodID, good.Name, info.Quantity, cost),
	}
	if err := tx.Create(scoreLog).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     exchange.ID,
		Status: exchange.Status,
	}
	return resp, nil
}

// List 获取兑奖记录
func (e *ExchangeSvc) List(params ExchangeListParams) (resp ExchangeForms, err error) {
	var exchanges []model.Exchange
	var total int64
	// 查询兑奖记录
	query := model.DB.Model(&model.Exchange{}).
		Where("user_id = ? AND status = ?", params.UserID, params.Status)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Scopes(model.Paginate(params.PagerForm)).
		Find(&exchanges).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Exchanges = make([]ExchangeForm, 0, len(exchanges))
	for _, ec := range exchanges {
		resp.Exchanges = append(resp.Exchanges, ExchangeForm{
			ID: ec.ID,
			Good: GoodForm{
				ID:        ec.Good.ID,
				Name:      ec.Good.Name,
				ThumbURL:  ec.Good.ThumbURL,
				NeedScore: ec.Good.NeedScore,
				Stock:     ec.Good.Stock,
			},
			Quantity:   ec.Quantity,
			ScoreCost:  ec.ScoreCost,
			Status:     ec.Status,
			ExchangeAt: ec.ExchangeAt,
			CreatedAt:  &ec.CreatedAt,
		})
	}

	return resp, nil
}
