package service

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type AdminActivitySvc struct{}

// List 获取活动列表（按开始时间倒序分页），支持 keyword 和 status 筛选
func (aa *AdminActivitySvc) List(params AdminActivityListParams) (resp ActivityCardPage, err error) {
	var total int64
	var activities []model.Activity

	query := model.DB.Model(&model.Activity{})

	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		if id, parseErr := strconv.ParseInt(params.Keyword, 10, 64); parseErr == nil {
			query = query.Where("id = ? OR title LIKE ? OR description LIKE ?", id, keyword, keyword)
		} else {
			query = query.Where("title LIKE ? OR description LIKE ?", keyword, keyword)
		}
	}

	now := time.Now()
	switch params.Status {
	case "not_started":
		query = query.Where("start_time > ?", now)
	case "active":
		query = query.Where("start_time <= ? AND end_time >= ?", now, now)
	case "ended":
		query = query.Where("end_time < ?", now)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	query = query.Order("start_time DESC")

	if err := query.Scopes(model.Paginate(params.PagerForm)).
		Find(&activities).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 批量统计每个活动的已审核题目数量
	photoCountMap := make(map[int64]int, len(activities))
	if len(activities) > 0 {
		actIDs := make([]int64, len(activities))
		for i, act := range activities {
			actIDs[i] = act.ID
		}
		type countResult struct {
			ActivityID int64
			Count      int
		}
		var counts []countResult
		model.DB.Model(&model.Photo{}).
			Select("activity_id, COUNT(*) as count").
			Where("activity_id IN ? AND status = ?", actIDs, "approved").
			Group("activity_id").
			Find(&counts)
		for _, c := range counts {
			photoCountMap[c.ActivityID] = c.Count
		}
	}

	resp.Total = total
	for _, activity := range activities {
		var startTime, endTime time.Time
		if activity.StartTime != nil {
			startTime = *activity.StartTime
		}
		if activity.EndTime != nil {
			endTime = *activity.EndTime
		}
		resp.List = append(resp.List, ActivityCard{
			ID:          activity.BaseModel.ID,
			Title:       activity.Title,
			CoverImage: Media{
				OriginURL:   activity.CoverURL,
				Width:       activity.CoverWidth,
				Height:      activity.CoverHeight,
			},
			Description: activity.Description,
			StartTime:   startTime,
			EndTime:     endTime,
			PhotoCount:  photoCountMap[activity.ID],
		})
	}

	return resp, nil
}

// Create 创建新活动
func (aa *AdminActivitySvc) Create(form AdminActivityCreate) (resp ResponseIS, err error) {
	// 校验时间
	if form.StartTime == nil || form.EndTime == nil {
		return resp, common.ErrNew(errors.New("活动时间不能为空"), common.ParamErr)
	}
	if !form.EndTime.After(*form.StartTime) {
		return resp, common.ErrNew(errors.New("结束时间必须晚于开始时间"), common.ParamErr)
	}

	// 上传封面图（必填）
	uploadResult, err := saveUploadedFile(form.CoverFile, "photos", false)
	if err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	activity := &model.Activity{
		Title:       form.Title,
		CoverURL:    uploadResult.ImageURL,
		CoverWidth:  uploadResult.ImageWidth,
		CoverHeight: uploadResult.ImageHeight,
		Description: form.Description,
		StartTime:   form.StartTime,
		EndTime:     form.EndTime,
		IsActive:    false,
	}

	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 保存答题奖励阶梯
	if err := saveRewardTiers(tx, activity.ID, form.RewardTiers); err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     activity.ID,
		Status: "success",
	}
	return resp, nil
}

// Update 更新活动信息
func (aa *AdminActivitySvc) Update(form AdminActivityUpdate) (resp ResponseIS, err error) {
	var activity model.Activity
	if err := model.DB.First(&activity, form.ActivityID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("活动不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 不允许修改已结束的活动
	if activity.EndTime != nil && activity.EndTime.Before(time.Now()) {
		return resp, common.ErrNew(errors.New("已结束的活动不可修改"), common.OpErr)
	}

	updates := map[string]any{}
	if form.Title != "" {
		updates["title"] = form.Title
	}
	if form.Description != "" {
		updates["description"] = form.Description
	}
	if form.StartTime != nil {
		updates["start_time"] = form.StartTime
	}
	if form.EndTime != nil {
		updates["end_time"] = form.EndTime
	}
	// 封面图可选，仅当提供时更新
	if form.CoverFile != nil {
		coverResult, uploadErr := saveUploadedFile(form.CoverFile, "photos", false)
		if uploadErr != nil {
			return resp, common.ErrNew(uploadErr, common.SysErr)
		}
		updates["cover_url"] = coverResult.ImageURL
		updates["cover_width"] = coverResult.ImageWidth
		updates["cover_height"] = coverResult.ImageHeight
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&activity).Updates(updates).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	// 更新答题奖励阶梯（提供时先删后建）
	if form.RewardTiers != "" {
		tx := model.DB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r)
			}
		}()

		if err := tx.Where("activity_id = ?", form.ActivityID).Delete(&model.AttemptRewardTier{}).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
		if err := saveRewardTiers(tx, form.ActivityID, form.RewardTiers); err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}

		if err := tx.Commit().Error; err != nil {
			return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
		}
	}

	resp = ResponseIS{
		ID:     activity.ID,
		Status: "success",
	}
	return resp, nil
}

// rewardTierInput 答题奖励阶梯输入结构
type rewardTierInput struct {
	Batch         int `json:"batch"`
	RankLimit     int `json:"rank_limit"`
	AttemptPoints int `json:"attempt_points"`
}

// saveRewardTiers 解析 JSON 并批量创建奖励阶梯
func saveRewardTiers(tx *gorm.DB, activityID int64, rewardTiersJSON string) error {
	if rewardTiersJSON == "" {
		return nil
	}
	var inputs []rewardTierInput
	if err := json.Unmarshal([]byte(rewardTiersJSON), &inputs); err != nil {
		return err
	}
	if len(inputs) == 0 {
		return nil
	}
	// 按 batch 排序以保证插入顺序
	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Batch < inputs[j].Batch
	})
	tiers := make([]model.AttemptRewardTier, 0, len(inputs))
	for _, in := range inputs {
		tiers = append(tiers, model.AttemptRewardTier{
			ActivityID:    activityID,
			Batch:         in.Batch,
			RankLimit:     in.RankLimit,
			AttemptPoints: in.AttemptPoints,
		})
	}
	return tx.Create(&tiers).Error
}
