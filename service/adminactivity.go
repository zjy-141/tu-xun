package service

import (
	"encoding/json"
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type AdminActivitySvc struct{}

// List 获取往期活动列表（按开始时间倒序分页）
func (aa *AdminActivitySvc) List(params AdminActivityListParams) (resp AdminActivityForms, err error) {

	var total int64
	var activitys []model.Activity

	query := model.DB.Model(&model.Activity{})

	if params.Keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+params.Keyword+"%", "%"+params.Keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	//按时间倒序排列
	query = query.Order("start_time DESC")

	if err := query.Scopes(model.Paginate(params.PagerForm)).
		Find(&activitys).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	for _, activity := range activitys {
		resp.Activities = append(resp.Activities, AdminActivityForm{
			ID:          activity.BaseModel.ID,
			Title:       activity.Title,
			CoverURL:    activity.CoverURL,
			Description: activity.Description,
			IsActive:    activity.IsActive,
			StartTime:   activity.StartTime,
			EndTime:     activity.EndTime,
		})
	}

	return resp, nil
}

// Detail 获取活动详情（含奖励阶梯配置）
func (aa *AdminActivitySvc) Detail(params AdminActivityGetByIDParams) (resp AdminActivityDetail, err error) {
	var activity model.Activity
	if err := model.DB.Preload("AttemptRewardTiers").
		First(&activity, params.ActivityID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("活动不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	tiers := make([]RewardTierInput, 0, len(activity.AttemptRewardTiers))
	for _, t := range activity.AttemptRewardTiers {
		tiers = append(tiers, RewardTierInput{
			Batch:         t.Batch,
			RankLimit:     t.RankLimit,
			AttemptPoints: t.AttemptPoints,
		})
	}

	resp = AdminActivityDetail{
		ID:          activity.BaseModel.ID,
		Title:       activity.Title,
		CoverURL:    activity.CoverURL,
		Description: activity.Description,
		IsActive:    activity.IsActive,
		StartTime:   activity.StartTime,
		EndTime:     activity.EndTime,
		PhotoPoints: activity.PhotoPoints,
		Tiers:       tiers,
	}

	return resp, nil
}

// Create 创建新活动（支持奖励阶梯配置）
func (aa *AdminActivitySvc) Create(info AdminActivityCreate) (resp ResponseIS, err error) {
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
	// 手动解析 RewardTiers
	var tiers []RewardTierInput
	if info.RewardTiers != "" {
		if err = json.Unmarshal([]byte(info.RewardTiers), &tiers); err != nil {
			// 处理 JSON 解析错误
			return resp, common.ErrNew(err, common.ParamErr)
		}
	}
	// 校验时间
	if info.StartTime == nil || info.EndTime == nil {
		return resp, common.ErrNew(errors.New("活动时间不能为空"), common.ParamErr)
	}
	if !info.EndTime.After(*info.StartTime) {
		return resp, common.ErrNew(errors.New("结束时间必须晚于开始时间"), common.ParamErr)
	}

	// 上传封面图
	imageURL := ""
	if info.CoverFile != nil {
		imageURL, _, err = saveUploadedFile(info.CoverFile, "photos", false)
		if err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	// 创建活动
	activity := &model.Activity{
		Title:       info.Title,
		CoverURL:    imageURL,
		Description: info.Description,
		StartTime:   info.StartTime,
		EndTime:     info.EndTime,
		IsActive:    false,
		PhotoPoints: *info.PhotoPoints,
	}

	if err := tx.Create(activity).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 创建奖励阶梯
	if len(tiers) > 0 {
		attemptRewardTier := make([]model.AttemptRewardTier, 0, len(info.RewardTiers))
		for _, t := range tiers {
			attemptRewardTier = append(attemptRewardTier, model.AttemptRewardTier{
				ActivityID:    activity.ID,
				Batch:         t.Batch,
				RankLimit:     t.RankLimit,
				AttemptPoints: t.AttemptPoints,
			})
		}
		if err := tx.Create(&attemptRewardTier).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     activity.ID,
		Status: "success",
	}
	return resp, nil
}

// Update 更新活动信息（支持奖励阶梯替换）
func (aa *AdminActivitySvc) Update(info AdminActivityUpdate) (resp ResponseIS, err error) {
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

	// 手动解析 RewardTiers
	var tiers []RewardTierInput
	if info.RewardTiers != "" {
		if err = json.Unmarshal([]byte(info.RewardTiers), &tiers); err != nil {
			// 处理 JSON 解析错误
			return resp, common.ErrNew(err, common.ParamErr)
		}
	}
	var activity model.Activity
	if err := tx.First(&activity, info.ActivityID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("活动不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	updates := map[string]interface{}{}
	if info.Title != "" {
		updates["title"] = info.Title
	}
	if info.Description != "" {
		updates["description"] = info.Description
	}
	if info.CoverFile != nil {
		coverURL, _, err := saveUploadedFile(info.CoverFile, "photos", false)
		if err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
		updates["cover_url"] = coverURL
	} else {
		var imageUrl string
		if err := tx.Model(&model.Photo{}).
			Where("activity_id = ?", info.ActivityID).
			Pluck("image_url", &imageUrl).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
		if imageUrl != "" {
			updates["cover_url"] = imageUrl
		}
	}
	if info.StartTime != nil {
		updates["start_time"] = info.StartTime
	}
	if info.EndTime != nil {
		updates["end_time"] = info.EndTime
	}
	if info.PhotoPoints != nil {
		updates["photo_points"] = info.PhotoPoints
	}

	if len(updates) > 0 {
		if err := tx.Model(&activity).Updates(updates).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	// 奖励阶梯：先删后建（替换策略）
	if tiers != nil {
		if err := tx.Where("activity_id = ?", info.ActivityID).Delete(&model.AttemptRewardTier{}).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
		if len(tiers) > 0 {
			attemptRewardTier := make([]model.AttemptRewardTier, 0, len(tiers))
			for _, t := range tiers {
				attemptRewardTier = append(attemptRewardTier, model.AttemptRewardTier{
					ActivityID:    info.ActivityID,
					Batch:         t.Batch,
					RankLimit:     t.RankLimit,
					AttemptPoints: t.AttemptPoints,
				})
			}
			if err := tx.Create(&attemptRewardTier).Error; err != nil {
				return resp, common.ErrNew(err, common.SysErr)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     activity.ID,
		Status: "success",
	}
	return resp, nil
}

// Notice 发布活动公告
func (aa *AdminActivitySvc) Notice(info AdminActivityNotice) (resp ResponseIS, err error) {
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

	var activity model.Activity
	if err := tx.First(&activity, info.ActivityID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("活动不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	notice := &model.Notice{
		Type:       "notice",
		Title:      info.Title,
		Content:    info.Content,
		ActivityID: info.ActivityID,
	}

	if err := tx.Create(notice).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     notice.ID,
		Status: "success",
	}
	return resp, nil
}
