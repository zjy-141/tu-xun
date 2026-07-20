package service

import (
	"errors"

	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type AdminGoodSvc struct{}

// ListGoods 管理员获取所有奖品列表（已分发/未分发）
func (ag *AdminGoodSvc) List(info AdminListGoodsParams) (resp AdminGoodForms, err error) {
	var goods []model.Good
	var total int64

	query := model.DB.Model(&model.Good{})

	if info.Available {
		query = query.Where(gorm.Expr("stock > ?", 0))
	}
	if info.Status != "" {
		query = query.Where("status = ?", info.Status)
	}
	if info.Keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+info.Keyword+"%", "%"+info.Keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at DESC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&goods).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Goods = make([]AdminGoodForm, 0, len(goods))
	for _, g := range goods {
		resp.Goods = append(resp.Goods, AdminGoodForm{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			ThumbURL:    g.ThumbURL,
			NeedScore:   g.NeedScore,
			Stock:       g.Stock,
			Status:      g.Status,
			CreatedAt:   g.BaseModel.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return resp, nil
}

// AdminGetByID 获取奖品详情
func (ag *AdminGoodSvc) GetByID(params AdminGoodGetByIDParams) (resp AdminGoodDetail, err error) {
	var good model.Good
	if err := model.DB.First(&good, params.GoodID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("奖品不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = AdminGoodDetail{
		ID:          good.ID,
		Name:        good.Name,
		Description: good.Description,
		ImageURL:    good.ImageURL,
		NeedScore:   good.NeedScore,
		Stock:       good.Stock,
		Status:      good.Status,
		CreatedAt:   good.BaseModel.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

// Create 新增商品
func (ag *AdminGoodSvc) Create(params GoodCreateParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 保存图片
	imageURL, thumbURL, err := saveUploadedFile(params.ImageFile, "goods", true)
	if err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
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

// Update 更新商品
func (ag *AdminGoodSvc) Update(params GoodUpdateParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	var good model.Good
	if err := tx.First(&good, params.GoodID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("奖品不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	updates := map[string]interface{}{}
	if params.Name != "" {
		updates["name"] = params.Name
	}
	if params.Description != "" {
		updates["description"] = params.Description
	}
	if params.NeedScore > 0 {
		updates["need_score"] = params.NeedScore
	}
	if params.Stock > 0 {
		updates["stock"] = params.Stock
	}
	if params.ImageFile != nil {
		imageURL, thumbURL, err := saveUploadedFile(params.ImageFile, "goods", true)
		if err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
		updates["image_url"] = imageURL
		updates["thumb_url"] = thumbURL
	}
	if params.Status != "" {
		updates["status"] = params.Status
	}

	if len(updates) > 0 {
		if err := tx.Model(&good).Updates(updates).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}
	resp = ResponseIS{
		ID:     good.ID,
		Status: good.Status,
	}
	return resp, nil
}

// Delete 删除商品
func (ag *AdminGoodSvc) Delete(params AdminGoodGetByIDParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var good model.Good
	if err := tx.First(&good, params.GoodID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("奖品不存在"), common.OpErr)
		}
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Delete(&good).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	resp = ResponseIS{
		ID:     good.ID,
		Status: "success",
	}
	return resp, nil
}

// Status 更新商品状态
func (ag *AdminGoodSvc) Status(params AdminGoodStatusParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	result := tx.Model(&model.Good{}).
		Where("id = ?", params.GoodID).
		Update("status", params.Status)

	// 处理更新错误
	if result.Error != nil {
		tx.Rollback()
		return resp, common.ErrNew(result.Error, common.SysErr)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("没有找到该商品"), common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	resp = ResponseIS{
		ID:     params.GoodID,
		Status: params.Status,
	}
	return resp, nil
}

// Stock 更新商品库存
func (ag *AdminGoodSvc) Stock(params GoodUpdateStockParams) (resp GoodStock, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var good model.Good
	if err := tx.First(&good, params.GoodID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("奖品不存在"), common.OpErr)
		}
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	good.Stock = params.Stock

	if err := tx.Save(&good).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	resp = GoodStock{
		ID:    good.ID,
		Stock: good.Stock,
	}
	return resp, nil
}

// // ClaimPrize 标记奖品已发放
// func (a *AdminSvc) ClaimPrize(prizeID int64) (resp AdminClaimPrizeResponse, err error) {
// 	tx := model.DB.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 			panic(r)
// 		}
// 	}()

// 	var prize model.Prize
// 	if err := tx.First(&prize, prizeID).Error; err != nil {
// 		tx.Rollback()
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return resp, common.ErrNew(errors.New("奖品记录不存在"), common.OpErr)
// 		}
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	if prize.Status == "claimed" {
// 		tx.Rollback()
// 		return resp, common.ErrNew(errors.New("该奖品已发放"), common.OpErr)
// 	}

// 	prize.Status = "claimed"
// 	if err := tx.Save(&prize).Error; err != nil {
// 		tx.Rollback()
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
// 	}

// 	resp = AdminClaimPrizeResponse{
// 		PrizeID: prize.ID,
// 		Status:  prize.Status,
// 	}
// 	return resp, nil
// }
