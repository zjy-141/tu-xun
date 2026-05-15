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
func (a *Admin) PendingPhotos(page, limit int) (map[string]any, error) {
	var photos []model.Photo
	var total int64

	query := model.DB.Model(&model.Photo{}).Where("status = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("Author").
		Order("created_at ASC").
		Scopes(model.Paginate(common.PagerForm{Page: page, Limit: limit})).
		Find(&photos).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	type PendingPhotoItem struct {
		ID             int64  `json:"id"`
		Title          string `json:"title"`
		LocationSecret string `json:"location_secret"`
		Author         struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"author"`
		CreatedAt time.Time `json:"created_at"`
	}

	items := make([]PendingPhotoItem, 0, len(photos))
	for _, ph := range photos {
		item := PendingPhotoItem{
			ID:             ph.ID,
			Title:          ph.Title,
			LocationSecret: ph.LocationSecret,
			CreatedAt:      ph.CreatedAt,
		}
		item.Author.ID = ph.Author.ID
		item.Author.Name = ph.Author.Name
		items = append(items, item)
	}

	return map[string]any{
		"total": total,
		"items": items,
	}, nil
}

// ReviewPhoto 审核图片
func (a *Admin) ReviewPhoto(photoID int64, action, rejectReason string) (map[string]any, error) {
	var photo model.Photo
	if err := model.DB.First(&photo, photoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	if photo.Status != "pending" {
		return nil, common.ErrNew(errors.New("该图片已审核过"), common.OpErr)
	}

	switch action {
	case "approve":
		photo.Status = "approved"
	case "reject":
		if rejectReason == "" {
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
	if action == "reject" {
		msg = "图片已拒绝: " + rejectReason
	}

	return map[string]any{
		"id":      photo.ID,
		"status":  photo.Status,
		"message": msg,
	}, nil
}

// PendingAttempts 获取待审核答题记录
func (a *Admin) PendingAttempts(page, limit int) (map[string]any, error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{}).Where("status = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("User").Preload("Photo").
		Order("created_at ASC").
		Scopes(model.Paginate(common.PagerForm{Page: page, Limit: limit})).
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
func (a *Admin) ReviewAttempt(attemptID int64, action, rejectReason string) (map[string]any, error) {
	var attempt model.Attempt
	if err := model.DB.Preload("Photo").First(&attempt, attemptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("答题记录不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	if attempt.Status != "pending" {
		return nil, common.ErrNew(errors.New("该答题记录已审核过"), common.OpErr)
	}

	now := time.Now()

	switch action {
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
		if rejectReason == "" {
			return nil, common.ErrNew(errors.New("拒绝时必须填写拒绝原因"), common.ParamErr)
		}
		attempt.Status = "rejected"
		attempt.RejectReason = rejectReason
		attempt.ReviewedAt = &now

	default:
		return nil, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	if err := model.DB.Save(&attempt).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	msg := "审核通过，恭喜答对！将为您发放纪念奖品。"
	if action == "reject" {
		msg = "审核未通过: " + rejectReason
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
