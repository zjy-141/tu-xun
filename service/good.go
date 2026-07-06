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
			return resp, errors.New("奖品不存在")
		}
		return resp, err
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

// Create 新增商品
func (s *GoodSvc) Create(params GoodCreateParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 保存图片
	imageURL, thumbURL, err := saveUploadedFile(params.ImageFile, "good")
	if err != nil {
		tx.Rollback()
		return resp, err
	}

	if params.Status == "" {
		params.Status = "inStore"
	}
	good := &model.Good{
		Name:        params.Name,
		Description: params.Description,
		NeedScore:   params.NeedScore,
		Stock:       params.Stock,
		ImageURL:    imageURL,
		ThumbURL:    thumbURL,
		Status:      params.Status,
	}

	if err := tx.Create(good).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	resp = ResponseIS{
		ID:     good.ID,
		Status: good.Status,
	}
	return resp, nil
}
