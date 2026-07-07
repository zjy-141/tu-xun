package service

import (
	"errors"

	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type GoodSvc struct{}

// List 获取奖品列表
func (s *GoodSvc) List(params GoodListParams) (resp GoodForms, err error) {
	var goods []model.Good
	var total int64

	query := model.DB.Model(&model.Good{}).Where("status = ?", "inStore")

	if params.Available {
		query = query.Where(gorm.Expr("stock > ?", 0))
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Scopes(model.Paginate(params.PagerForm)).
		Order("id DESC").
		Find(&goods).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	goodForms := make([]GoodForm, 0, len(goods))
	for _, g := range goods {
		goodForms = append(goodForms, GoodForm{
			ID:        g.ID,
			Name:      g.Name,
			ThumbURL:  g.ThumbURL,
			NeedScore: g.NeedScore,
			Stock:     g.Stock,
		})
	}

	resp = GoodForms{
		Total: total,
		Goods: goodForms,
	}
	return resp, nil
}

// GetByID 获取奖品详情
func (s *GoodSvc) GetByID(params GoodGetByIDParams) (resp GoodDetail, err error) {
	var good model.Good
	if err := model.DB.Where("status = ?", "inStore").First(&good, params.GoodID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("奖品不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = GoodDetail{
		ID:          good.ID,
		Name:        good.Name,
		Description: good.Description,
		ImageURL:    good.ImageURL,
		NeedScore:   good.NeedScore,
		Stock:       good.Stock,
	}

	return resp, nil
}
