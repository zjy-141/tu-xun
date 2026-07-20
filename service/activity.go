package service

import (
	"errors"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type ActivitySvc struct{}

// Current 获取当前活动
func (a *ActivitySvc) Current() (resp ActivityForm, err error) {

	now := time.Now().UTC()
	var activity model.Activity
	if err := model.DB.Where("(start_time <= ? AND end_time >= ?) OR is_active = ?", now, now, true).
		First(&activity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("当前没有活动开放"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = ActivityForm{
		ID:          activity.BaseModel.ID,
		Title:       activity.Title,
		CoverURL:    activity.CoverURL,
		Description: activity.Description,
		IsActive:    activity.IsActive,
		StartTime:   activity.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:     activity.EndTime.Format("2006-01-02 15:04:05"),
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
		resp.Activities = append(resp.Activities, ActivityForm{
			ID:          activity.BaseModel.ID,
			Title:       activity.Title,
			CoverURL:    activity.CoverURL,
			Description: activity.Description,
			IsActive:    activity.IsActive,
			StartTime:   activity.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:     activity.EndTime.Format("2006-01-02 15:04:05"),
		})
	}

	return resp, nil
}
