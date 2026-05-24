package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Attempt struct{}

// Create 提交答题
func (a *Attempt) Create(info CreateAttemptParams) (*model.Attempt, error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 检查图片是否存在且已审核通过
	var photo model.Photo
	if err := tx.First(&photo, info.PhotoID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}
	if photo.Status != "approved" {
		tx.Rollback()
		return nil, common.ErrNew(errors.New("该图片尚未通过审核，暂不可答题"), common.OpErr)
	}

	// 检查是否已有待审核的答题记录（同一用户同一图片）
	var existAttempt model.Attempt
	if err := tx.Where("photo_id = ? AND user_id = ? AND status = ?", info.PhotoID, info.UserID, "pending").
		First(&existAttempt).Error; err == nil {
		tx.Rollback()
		return nil, common.ErrNew(errors.New("您已提交过答题，请等待管理员审核"), common.OpErr)
	}

	// 保存答题图片
	imageURL, _, err := saveUploadedFile(info.ImageFile, "attempts")
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	attempt := &model.Attempt{
		PhotoID:         info.PhotoID,
		UserID:          info.UserID,
		ImageURL:        imageURL,
		GuessedLocation: info.GuessedLocation,
		Status:          "pending",
		IsWinner:        false,
	}

	if err := tx.Create(attempt).Error; err != nil {
		tx.Rollback()
		return nil, common.ErrNew(err, common.SysErr)
	}

	// 更新图片的答题次数
	tx.Model(&photo).UpdateColumn("attempts_count", gorm.Expr("attempts_count + 1"))

	if err := tx.Commit().Error; err != nil {
		return nil, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return attempt, nil
}

// MyAttempts 获取我对某图片的所有答题记录
func (a *Attempt) MyAttempts(info MyAttemptsParams) (resp MyAttemptsResponse, err error) {
	var photo model.Photo
	if err := model.DB.First(&photo, info.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	var attempts []model.Attempt
	if err := model.DB.Where("photo_id = ? AND user_id = ?", info.PhotoID, info.UserID).
		Order("created_at DESC").Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
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

	resp = MyAttemptsResponse{
		PhotoID:    info.PhotoID,
		Solved:     photo.Solved,
		MyAttempts: items,
	}
	return resp, nil
}
