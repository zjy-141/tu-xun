package service

import (
	"time"
	"tu-xun/common"
	"tu-xun/model"
)

type ActivitySvc struct{}

// Active 获取当前进行中的活动列表（分页）
func (a *ActivitySvc) Active(params common.PagerForm) (resp ActivityForms, err error) {
	now := time.Now()
	var total int64
	var activities []model.Activity

	query := model.DB.Model(&model.Activity{}).
		Where("start_time <= ? AND end_time > ?", now, now)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("start_time DESC").
		Scopes(model.Paginate(params)).
		Find(&activities).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	for _, activity := range activities {
		resp.List = append(resp.List, ActivityForm{
			ID:          activity.BaseModel.ID,
			Title:       activity.Title,
			CoverURL:    activity.CoverURL,
			Description: activity.Description,
			StartTime:   *activity.StartTime,
			EndTime:     *activity.EndTime,
		})
	}

	return resp, nil
}

// History 获取往期活动列表（按开始时间倒序分页）
func (a *ActivitySvc) History(params common.PagerForm) (resp ActivityForms, err error) {

	var total int64
	var activitys []model.Activity

	query := model.DB.Model(&model.Activity{})

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	//按时间倒序排列
	query = query.Order("start_time DESC")

	if err := query.Scopes(model.Paginate(params)).
		Find(&activitys).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	for _, activity := range activitys {
		var startTime, endTime time.Time
		if activity.StartTime != nil {
			startTime = *activity.StartTime
		}
		if activity.EndTime != nil {
			endTime = *activity.EndTime
		}
		resp.List = append(resp.List, ActivityForm{
			ID:          activity.BaseModel.ID,
			Title:       activity.Title,
			CoverURL:    activity.CoverURL,
			Description: activity.Description,
			StartTime:   startTime,
			EndTime:     endTime,
		})
	}

	return resp, nil
}

// IsActive 检查活动是否正在进行中
func (a *ActivitySvc) IsActivityActive(activityID int64) (bool, error) {
	now := time.Now()
	var count int64
	if err := model.DB.Model(&model.Activity{}).
		Where("id = ? AND start_time <= ? AND end_time > ?", activityID, now, now).
		Count(&count).Error; err != nil {
		return false, common.ErrNew(err, common.SysErr)
	}
	return count > 0, nil
}
