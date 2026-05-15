package service

import (
	"errors"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Attempt struct{}

// Create 提交答题
func (a *Attempt) Create(p CreateAttemptParams) (*model.Attempt, error) {
	// 检查图片是否存在且已审核通过
	var photo model.Photo
	if err := model.DB.First(&photo, p.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}
	if photo.Status != "approved" {
		return nil, common.ErrNew(errors.New("该图片尚未通过审核，暂不可答题"), common.OpErr)
	}

	// 检查是否已有待审核的答题记录（同一用户同一图片）
	var existAttempt model.Attempt
	if err := model.DB.Where("photo_id = ? AND user_id = ? AND status = ?", p.PhotoID, p.UserID, "pending").
		First(&existAttempt).Error; err == nil {
		return nil, common.ErrNew(errors.New("您已提交过答题，请等待管理员审核"), common.OpErr)
	}

	// 保存答题图片
	imageURL, _, err := saveUploadedFile(p.ImageFile, "attempts")
	if err != nil {
		return nil, err
	}

	attempt := &model.Attempt{
		PhotoID:         p.PhotoID,
		UserID:          p.UserID,
		ImageURL:        imageURL,
		GuessedLocation: p.GuessedLocation,
		Status:          "pending",
		IsWinner:        false,
	}

	if err := model.DB.Create(attempt).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	// 更新图片的答题次数
	model.DB.Model(&photo).UpdateColumn("attempts_count", gorm.Expr("attempts_count + 1"))

	return attempt, nil
}

// MyAttempts 获取我对某图片的所有答题记录
func (a *Attempt) MyAttempts(p MyAttemptsParams) (map[string]any, error) {
	var photo model.Photo
	if err := model.DB.First(&photo, p.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	var attempts []model.Attempt
	if err := model.DB.Where("photo_id = ? AND user_id = ?", p.PhotoID, p.UserID).
		Order("created_at DESC").Find(&attempts).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	type AttemptItem struct {
		ID              int64      `json:"id"`
		ImageURL        string     `json:"image_url"`
		GuessedLocation string     `json:"guessed_location"`
		Status          string     `json:"status"`
		IsWinner        bool       `json:"is_winner"`
		ReviewedAt      *time.Time `json:"reviewed_at"`
	}

	items := make([]AttemptItem, 0, len(attempts))
	for _, at := range attempts {
		items = append(items, AttemptItem{
			ID:              at.ID,
			ImageURL:        at.ImageURL,
			GuessedLocation: at.GuessedLocation,
			Status:          at.Status,
			IsWinner:        at.IsWinner,
			ReviewedAt:      at.ReviewedAt,
		})
	}

	return map[string]any{
		"photo_id":    p.PhotoID,
		"solved":      photo.Solved,
		"my_attempts": items,
	}, nil
}
