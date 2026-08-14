package service

import (
	"errors"
	"fmt"
	"tu-xun/common"
	"tu-xun/model"
	"tu-xun/pkg/urlutil"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminExchangeSvc struct{}

// List 获取兑奖记录（管理端，查看所有用户）
func (ae *AdminExchangeSvc) List(params AdminExchangeListParams) (resp AdminExchangePage, err error) {
	var exchanges []model.Exchange
	var total int64

	query := model.DB.Model(&model.Exchange{})

	if params.Status != "" {
		query = query.Where("exchange.status = ?", params.Status)
	}
	if params.Keyword != "" {
		query = query.Where("CAST(exchange.id AS CHAR) LIKE ?", "%"+params.Keyword+"%")
	}
	if params.UserKeyword != "" {
		query = query.Joins("JOIN user ON user.id = exchange.user_id AND user.nickname LIKE ?", "%"+params.UserKeyword+"%")
	}
	if params.GoodKeyword != "" {
		query = query.Joins("JOIN good ON good.id = exchange.good_id AND good.name LIKE ?", "%"+params.GoodKeyword+"%")
	}
	if params.VerifyCode != "" {
		query = query.Where("exchange.verify_code = ?", params.VerifyCode)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("Good").Preload("User").
		Scopes(model.Paginate(params.PagerForm)).
		Order("exchange.created_at DESC").
		Find(&exchanges).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]AdminExchangeItem, 0, len(exchanges))
	for _, ec := range exchanges {
		resp.List = append(resp.List, AdminExchangeItem{
			ID: ec.ID,
			User: UserBrief{
				ID:        ec.User.ID,
				Nickname:  ec.User.Nickname,
				Avatar: urlutil.FullURL(ec.User.AvatarURL),
			},
			Good: GoodBrief{
				ID:         ec.Good.ID,
				Name:       ec.Good.Name,
				Image: Media{ThumbURL: urlutil.FullURL(ec.Good.ThumbURL), Width: ec.Good.ThumbWidth, Height: ec.Good.ThumbHeight},
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

// Verify 管理端核销/取消兑奖记录
func (ae *AdminExchangeSvc) Verify(params AdminExchangeVerifyParams) error {
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

	// 1. 锁定兑换记录
	var exchange model.Exchange
	if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", params.ExchangeID).
		First(&exchange).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNew(errors.New("兑换记录不存在"), common.ParamErr)
		}
		return common.ErrNew(err, common.SysErr)
	}

	// 2. 状态校验：只有 pending 状态才能操作
	if exchange.Status != "pending" {
		tx.Rollback()
		return common.ErrNew(errors.New("该兑换记录已处理，无法重复操作"), common.ParamErr)
	}

	switch params.Action {
	case "verify":
		// 核销：更新状态为 verified，记录取货时间
		if err = tx.Model(&exchange).Updates(map[string]interface{}{
			"status":      "verified",
			"exchange_at": gorm.Expr("NOW()"),
		}).Error; err != nil {
			return common.ErrNew(err, common.SysErr)
		}
	case "cancel":
		// 取消：回退库存和积分，更新状态为 cancelled
		// 回退库存
		if err = tx.Model(&model.Good{}).
			Where("id = ?", exchange.GoodID).
			Update("stock", gorm.Expr("stock + ?", exchange.Quantity)).Error; err != nil {
			return common.ErrNew(err, common.SysErr)
		}
		scoreParams := ScoreChangeParams{
			UserID:      exchange.UserID,
			Delta:       exchange.ScoreCost,
			Reason:      fmt.Sprintf("管理员取消兑换记录 #%d，退回积分 %d", exchange.ID, exchange.ScoreCost),
			RelatedID:   exchange.ID,
			RelatedType: "exchange_cancel",
		}
		scoreSvc := ScoreSvc{}
		if _, scoreErr := scoreSvc.RegularScoreChange(tx, scoreParams); scoreErr != nil {
			tx.Rollback()
			return common.ErrNew(scoreErr, common.SysErr)
		}

		// 更新状态
		if err = tx.Model(&exchange).Update("status", "cancelled").Error; err != nil {
			return common.ErrNew(err, common.SysErr)
		}
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}

	return nil
}
