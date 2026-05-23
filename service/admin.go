package service

import (
	"errors"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Admin struct{}

// PendingPhotos 获取待审核图片列表
func (a *Admin) PendingPhotos(info common.PagerForm) (resp PendingPhotosResponse, err error) {
	var photos []model.Photo
	var total int64

	query := model.DB.Model(&model.Photo{}).Where("status = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("Author").Order("created_at ASC").
		Scopes(model.Paginate(info)).
		Find(&photos).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Photos = make([]PendingPhotoForm, 0, len(photos))
	for _, photo := range photos {
		resp.Photos = append(resp.Photos, PendingPhotoForm{
			ID:             photo.ID,
			Title:          photo.Title,
			Description:    photo.Description,
			LocationSecret: photo.LocationSecret,
			ThumbURL:       photo.ThumbURL,
		})
	}
	return resp, nil
}

// ReviewPhoto 审核图片
func (a *Admin) ReviewPhoto(p ReviewPhotoParams) (map[string]any, error) {
	var photo model.Photo
	if err := model.DB.First(&photo, p.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	if photo.Status != "pending" {
		return nil, common.ErrNew(errors.New("该图片已审核过"), common.OpErr)
	}

	switch p.Action {
	case "approve":
		photo.Status = "approved"
	case "reject":
		if p.RejectReason == "" {
			return nil, common.ErrNew(errors.New("拒绝时必须填写拒绝原因"), common.ParamErr)
		}
		photo.Status = "rejected"
	default:
		return nil, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	if err := model.DB.Save(&photo).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	msg := "图片已通过审核，现已公开"
	if p.Action == "reject" {
		msg = "图片已拒绝: " + p.RejectReason
	}

	return map[string]any{
		"id":      photo.ID,
		"status":  photo.Status,
		"message": msg,
	}, nil
}

// PendingAttempts 获取待审核答题记录
func (a *Admin) PendingAttempts(p PendingAttemptsParams) (map[string]any, error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{}).Where("status = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("User").Preload("Photo").
		Order("created_at ASC").
		Scopes(model.Paginate(common.PagerForm{Page: p.Page, Limit: p.Limit})).
		Find(&attempts).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	type PendingAttemptItem struct {
		AttemptID       int64     `json:"attempt_id"`
		PhotoID         int64     `json:"photo_id"`
		PhotoTitle      string    `json:"photo_title"`
		User            UserBrief `json:"user"`
		ImageURL        string    `json:"image_url"`
		GuessedLocation string    `json:"guessed_location"`
		SubmittedAt     time.Time `json:"submitted_at"`
	}

	items := make([]PendingAttemptItem, 0, len(attempts))
	for _, at := range attempts {
		items = append(items, PendingAttemptItem{
			AttemptID:       at.ID,
			PhotoID:         at.PhotoID,
			PhotoTitle:      at.Photo.Title,
			User:            UserBrief{ID: at.User.ID, Name: at.User.Name},
			ImageURL:        at.ImageURL,
			GuessedLocation: at.GuessedLocation,
			SubmittedAt:     at.CreatedAt,
		})
	}

	return map[string]any{
		"total": total,
		"items": items,
	}, nil
}

// ReviewAttempt 审核答题记录
func (a *Admin) ReviewAttempt(p ReviewAttemptParams) (map[string]any, error) {
	var attempt model.Attempt
	if err := model.DB.Preload("Photo").First(&attempt, p.AttemptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("答题记录不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	if attempt.Status != "pending" {
		return nil, common.ErrNew(errors.New("该答题记录已审核过"), common.OpErr)
	}

	now := time.Now()

	switch p.Action {
	case "approve":
		attempt.Status = "approved"
		attempt.ReviewedAt = &now

		// 自动判断该题目是否已被破解
		if attempt.Photo.Solved {
			// 已被破解 → 标记为通过但不获奖
			attempt.IsWinner = false
		} else {
			// 尚未被破解 → 标记获奖，并更新图片状态
			attempt.IsWinner = true
			model.DB.Model(&model.Photo{}).Where("id = ?", attempt.PhotoID).Update("solved", true)

			// 生成奖品记录
			prize := &model.Prize{
				UserID:    attempt.UserID,
				PhotoID:   attempt.PhotoID,
				PrizeType: "明信片套装",
				Status:    "unclaimed",
				AwardedAt: &now,
			}
			model.DB.Create(prize)

			// 更新用户获奖次数
			model.DB.Model(&model.User{}).Where("id = ?", attempt.UserID).
				UpdateColumn("prize_count", gorm.Expr("prize_count + 1"))
		}

	case "reject":
		if p.RejectReason == "" {
			return nil, common.ErrNew(errors.New("拒绝时必须填写拒绝原因"), common.ParamErr)
		}
		attempt.Status = "rejected"
		attempt.RejectReason = p.RejectReason
		attempt.ReviewedAt = &now

	default:
		return nil, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	if err := model.DB.Save(&attempt).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	msg := "审核通过，恭喜答对！将为您发放纪念奖品。"
	if p.Action == "reject" {
		msg = "审核未通过: " + p.RejectReason
	} else if !attempt.IsWinner {
		msg = "正确答案，但奖品已被领走。感谢您的参与！"
	}

	return map[string]any{
		"attempt_id":   attempt.ID,
		"status":       attempt.Status,
		"is_winner":    attempt.IsWinner,
		"photo_solved": attempt.Photo.Solved || attempt.IsWinner,
		"message":      msg,
	}, nil
}

// ClaimPrize 标记奖品已发放
func (a *Admin) ClaimPrize(prizeID int64) (map[string]any, error) {
	var prize model.Prize
	if err := model.DB.First(&prize, prizeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("奖品记录不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	if prize.Status == "claimed" {
		return nil, common.ErrNew(errors.New("该奖品已发放"), common.OpErr)
	}

	prize.Status = "claimed"
	if err := model.DB.Save(&prize).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	return map[string]any{
		"prize_id": prize.ID,
		"status":   prize.Status,
	}, nil
}
