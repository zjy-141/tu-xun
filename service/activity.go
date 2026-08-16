package service

import (
	"time"

	"tu-xun/common"
	"tu-xun/model"
	"tu-xun/pkg/urlutil"
)

type ActivitySvc struct{}

// List 获取活动卡片列表，排除未开始的活动，支持状态筛选和关键词搜索
func (a *ActivitySvc) List(params ActivityListParams) (ActivityCardPage, error) {
	now := time.Now()
	var total int64
	var activities []model.Activity

	// 排除未开始的活动：已开始 或 已结束
	query := model.DB.Model(&model.Activity{}).
		Where("start_time <= ?", now)

	// 状态筛选（永久活动不受 end_time 影响，始终视为进行中）
	switch params.Status {
	case "active":
		query = query.Where("is_active = ? OR end_time > ?", true, now)
	case "ended":
		query = query.Where("is_active = ? AND end_time <= ?", false, now)
	}

	// 关键词搜索（标题、描述）
	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", kw, kw)
	}

	if err := query.Count(&total).Error; err != nil {
		return ActivityCardPage{}, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("start_time DESC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&activities).Error; err != nil {
		return ActivityCardPage{}, common.ErrNew(err, common.SysErr)
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

	resp := ActivityCardPage{
		Total: total,
		List:  make([]ActivityCard, 0, len(activities)),
	}
	for _, act := range activities {
		var startTime, endTime time.Time
		if act.StartTime != nil {
			startTime = *act.StartTime
		}
		if act.EndTime != nil {
			endTime = *act.EndTime
		}
		resp.List = append(resp.List, ActivityCard{
			ID:          act.ID,
			Title:       act.Title,
			CoverImage: Media{
				OriginURL:   urlutil.FullURL(act.CoverURL),
				Width:       act.CoverWidth,
				Height:      act.CoverHeight,
			},
			Description: act.Description,
			StartTime:   startTime,
			EndTime:     endTime,
			PhotoCount:  photoCountMap[act.ID],
		})
	}

	return resp, nil
}

// IsActivityActive 检查活动是否正在进行中
func (a *ActivitySvc) IsActivityActive(activityID int64) (bool, error) {
	now := time.Now()
	var count int64
	if err := model.DB.Model(&model.Activity{}).
		Where("id = ? AND start_time <= ? AND (is_active = ? OR end_time > ?)", activityID, now, true, now).
		Count(&count).Error; err != nil {
		return false, common.ErrNew(err, common.SysErr)
	}
	return count > 0, nil
}
