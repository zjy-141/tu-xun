package service

import (
	"errors"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type AdminActivitySvc struct{}

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

	// 解析时间
	startTime, err := time.Parse("2006-01-02 15:04:05", info.StartTime)
	if err != nil {
		return resp, common.ErrNew(errors.New("活动开始时间格式错误"), common.ParamErr)
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", info.EndTime)
	if err != nil {
		return resp, common.ErrNew(errors.New("活动结束时间格式错误"), common.ParamErr)
	}
	if !endTime.After(startTime) {
		return resp, common.ErrNew(errors.New("结束时间必须晚于开始时间"), common.ParamErr)
	}

	// 上传封面图
	imageURL := ""
	if info.CoverFile != nil {
		imageURL, _, err = saveUploadedFile(info.CoverFile, "photos")
		if err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	// 创建活动
	activity := &model.Activity{
		Title:       info.Title,
		CoverURL:    imageURL,
		Description: info.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		IsActive:    false,
		PhotoPoints: *info.PhotoPoints,
	}

	if err := tx.Create(activity).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 创建奖励阶梯
	if len(info.RewardTiers) > 0 {
		tiers := make([]model.AttemptRewardTier, 0, len(info.RewardTiers))
		for _, t := range info.RewardTiers {
			tiers = append(tiers, model.AttemptRewardTier{
				ActivityID:    activity.ID,
				Batch:         t.Batch,
				RankLimit:     t.RankLimit,
				AttemptPoints: t.AttemptPoints,
			})
		}
		if err := tx.Create(&tiers).Error; err != nil {
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
		coverURL, _, err := saveUploadedFile(info.CoverFile, "photos")
		if err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
		updates["cover_url"] = coverURL
	}
	if info.StartTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", info.StartTime)
		if err != nil {
			return resp, common.ErrNew(errors.New("开始时间格式错误"), common.ParamErr)
		}
		updates["start_time"] = t
	}
	if info.EndTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", info.EndTime)
		if err != nil {
			return resp, common.ErrNew(errors.New("结束时间格式错误"), common.ParamErr)
		}
		updates["end_time"] = t
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
	if info.RewardTiers != nil {
		if err := tx.Where("activity_id = ?", info.ActivityID).Delete(&model.AttemptRewardTier{}).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
		if len(info.RewardTiers) > 0 {
			tiers := make([]model.AttemptRewardTier, 0, len(info.RewardTiers))
			for _, t := range info.RewardTiers {
				tiers = append(tiers, model.AttemptRewardTier{
					ActivityID:    info.ActivityID,
					Batch:         t.Batch,
					RankLimit:     t.RankLimit,
					AttemptPoints: t.AttemptPoints,
				})
			}
			if err := tx.Create(&tiers).Error; err != nil {
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
