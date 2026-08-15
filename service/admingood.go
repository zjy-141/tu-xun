package service

import (
	"errors"
	"strconv"
	"tu-xun/common"
	"tu-xun/model"
	"tu-xun/pkg/urlutil"

	"gorm.io/gorm"
)

type AdminGoodSvc struct{}

// List 管理员获取奖品列表，支持 status 和 keyword 筛选
func (ag *AdminGoodSvc) List(params AdminGoodListParams) (resp GoodItemPage, err error) {
	var goods []model.Good
	var total int64

	query := model.DB.Model(&model.Good{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		if id, parseErr := strconv.ParseInt(params.Keyword, 10, 64); parseErr == nil {
			query = query.Where("id = ? OR name LIKE ? OR description LIKE ?", id, keyword, keyword)
		} else {
			query = query.Where("name LIKE ? OR description LIKE ?", keyword, keyword)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at DESC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&goods).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]GoodItem, 0, len(goods))
	for _, g := range goods {
		resp.List = append(resp.List, GoodItem{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Image: Media{
				OriginURL:   urlutil.FullURL(g.ImageURL),
				ThumbURL:    urlutil.FullURL(g.ThumbURL),
				Width:       g.ImageWidth,
				Height:      g.ImageHeight,
			},
			ScorePrice:  g.NeedScore,
			Stock:       g.Stock,
			Status:      g.Status,
			CreatedAt:   &g.BaseModel.CreatedAt,
		})
	}

	return resp, nil
}

// Create 新增商品
func (ag *AdminGoodSvc) Create(form GoodCreateParams) (resp ResponseIS, err error) {
	// 保存图片
	uploadResult, err := saveUploadedFile(form.ImageFile, "goods", true)
	if err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if form.Status == "" {
		form.Status = "in_store"
	}

	good := &model.Good{
		Name:        form.Name,
		Description: form.Description,
		NeedScore:   *form.NeedScore,
		Stock:       *form.Stock,
		ImageURL:    uploadResult.ImageURL,
		ThumbURL:    uploadResult.ThumbURL,
		ImageWidth:  uploadResult.ImageWidth,
		ImageHeight: uploadResult.ImageHeight,
		ThumbWidth:  uploadResult.ThumbWidth,
		ThumbHeight: uploadResult.ThumbHeight,
		Status:      form.Status,
	}

	if err := model.DB.Create(good).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = ResponseIS{
		ID:     good.ID,
		Status: good.Status,
	}
	return resp, nil
}

// Update 更新商品（合并 status/stock 到主更新）
func (ag *AdminGoodSvc) Update(form GoodUpdateParams) (resp ResponseIS, err error) {
	var good model.Good
	if err := model.DB.First(&good, form.GoodID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("奖品不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	updates := map[string]interface{}{}
	if form.Name != "" {
		updates["name"] = form.Name
	}
	if form.Description != "" {
		updates["description"] = form.Description
	}
	if form.NeedScore != nil {
		updates["need_score"] = *form.NeedScore
	}
	if form.Stock != nil {
		updates["stock"] = *form.Stock
	}
	if form.Status != "" {
		updates["status"] = form.Status
	}
	if form.ImageFile != nil {
		uploadResult, uploadErr := saveUploadedFile(form.ImageFile, "goods", true)
		if uploadErr != nil {
			return resp, common.ErrNew(uploadErr, common.SysErr)
		}
		updates["image_url"] = uploadResult.ImageURL
		updates["thumb_url"] = uploadResult.ThumbURL
		updates["image_width"] = uploadResult.ImageWidth
		updates["image_height"] = uploadResult.ImageHeight
		updates["thumb_width"] = uploadResult.ThumbWidth
		updates["thumb_height"] = uploadResult.ThumbHeight
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&good).Updates(updates).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	resp = ResponseIS{
		ID:     good.ID,
		Status: good.Status,
	}
	return resp, nil
}

// Delete 删除商品（有兑换记录则拒绝）
func (ag *AdminGoodSvc) Delete(goodID int64) error {
	var count int64
	if err := model.DB.Model(&model.Exchange{}).Where("good_id = ?", goodID).Count(&count).Error; err != nil {
		return common.ErrNew(err, common.SysErr)
	}
	if count > 0 {
		return common.ErrNew(errors.New("已有兑换记录，不可删除"), common.OpErr)
	}

	result := model.DB.Delete(&model.Good{}, goodID)
	if result.Error != nil {
		return common.ErrNew(result.Error, common.SysErr)
	}
	if result.RowsAffected == 0 {
		return common.ErrNew(errors.New("奖品不存在"), common.OpErr)
	}

	return nil
}
