package service

import (
	"tu-xun/model"
)

// StatsSvc 管理端工作台统计
type StatsSvc struct{}

// GetStats 获取工作台统计数据
func (s *StatsSvc) GetStats() (AdminStats, error) {
	var stats AdminStats

	if err := model.DB.Model(&model.User{}).Where("status = ?", "active").Count(&stats.UserCount).Error; err != nil {
		return stats, err
	}
	if err := model.DB.Model(&model.Photo{}).Where("status = ?", "pending").Count(&stats.PendingPhotoCount).Error; err != nil {
		return stats, err
	}
	if err := model.DB.Model(&model.Attempt{}).Where("status = ?", "pending").Count(&stats.PendingAttemptCount).Error; err != nil {
		return stats, err
	}
	if err := model.DB.Model(&model.Comment{}).Where("status = ?", "pending").Count(&stats.PendingCommentCount).Error; err != nil {
		return stats, err
	}
	if err := model.DB.Model(&model.Feedback{}).Where("status = ?", "pending").Count(&stats.PendingFeedbackCount).Error; err != nil {
		return stats, err
	}

	return stats, nil
}
