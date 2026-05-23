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
func (a *Admin) PendingPhotos(info PendingPhotoParams) (resp PendingPhotosResponse, err error) {
	var photos []model.Photo
	var total int64
	query := model.DB.Model(&model.Photo{})
	if info.AdminLevel < 2 {
		// 普通管理员只能看到待审核的图片
		query = query.Where("status = ?", "pending")
	} else {
		query = query.Where("status = ?", info.status)
	}
	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at ASC").
		Scopes(model.Paginate(info.PagerForm)).
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
func (a *Admin) ReviewPhoto(info ReviewPhotoParams) (resp ReviewPhotoResponse, err error) {
	var photo model.Photo
	if err := model.DB.First(&photo, info.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if photo.Status != "pending" && info.AdminLevel < 2 {
		return resp, common.ErrNew(errors.New("该图片已审核过,请联系更高级管理员修改"), common.OpErr)
	}

	switch info.Action {
	case "approve":
		photo.Status = "approved"
	case "reject":
		if info.RejectReason == "" {
			return resp, common.ErrNew(errors.New("拒绝时请填写拒绝原因"), common.ParamErr)
		}
		photo.Status = "rejected"
	default:
		return resp, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	if err := model.DB.Save(&photo).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.ID = photo.ID
	resp.Status = photo.Status
	if info.Action == "approve" {
		resp.Message = "图片已通过审核，现已公开"
	} else if info.Action == "reject" {
		resp.Message = "图片已拒绝: " + info.RejectReason
	}
	return resp, nil
}

// PendingAttempts 获取待审核答题记录
func (a *Admin) PendingAttempts(info PendingAttemptParams) (resp PendingAttemptsResponse, err error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{})

	if info.AdminLevel < 2 {
		// 普通管理员只能看到待审核的答题记录
		query = query.Where("status = ?", "pending")
	} else {
		query = query.Where("status = ?", info.status)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("Photo").
		Order("created_at ASC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Attempts = make([]PendingAttemptForm, 0, len(attempts))
	for _, at := range attempts {
		resp.Attempts = append(resp.Attempts, PendingAttemptForm{
			AttemptID:       at.ID,
			PhotoID:         at.PhotoID,
			PhotoTitle:      at.Photo.Title,
			ImageURL:        at.ImageURL,
			GuessedLocation: at.GuessedLocation,
			SubmittedAt:     at.CreatedAt,
		})
	}
	return resp, nil
}

// ReviewAttempt 审核答题记录
func (a *Admin) ReviewAttempt(info ReviewAttemptParams) (resp ReviewAttemptResponse, err error) {
	var attempt model.Attempt
	if err := model.DB.Preload("Photo").First(&attempt, info.AttemptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("答题记录不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if attempt.Status != "pending" {
		return resp, common.ErrNew(errors.New("该答题记录已审核过"), common.OpErr)
	}

	now := time.Now()

	switch info.Action {
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
		if info.RejectReason == "" {
			return resp, common.ErrNew(errors.New("拒绝时请填写拒绝原因"), common.ParamErr)
		}
		attempt.Status = "rejected"
		attempt.RejectReason = info.RejectReason
		attempt.ReviewedAt = &now

	default:
		return resp, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	if err := model.DB.Save(&attempt).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	msg := "审核通过，恭喜答对！将为您发放纪念奖品。"
	if info.Action == "reject" {
		msg = "审核未通过: " + info.RejectReason
	} else if !attempt.IsWinner {
		msg = "正确答案，但奖品已被领走。感谢您的参与！"
	}

	resp = ReviewAttemptResponse{
		AttemptID:   attempt.ID,
		Status:      attempt.Status,
		IsWinner:    attempt.IsWinner,
		PhotoSolved: attempt.Photo.Solved || attempt.IsWinner,
		Message:     msg,
	}
	return resp, nil
}

// ClaimPrize 标记奖品已发放
func (a *Admin) ClaimPrize(prizeID int64) (resp ClaimPrizeResponse, err error) {
	var prize model.Prize
	if err := model.DB.First(&prize, prizeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("奖品记录不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if prize.Status == "claimed" {
		return resp, common.ErrNew(errors.New("该奖品已发放"), common.OpErr)
	}

	prize.Status = "claimed"
	if err := model.DB.Save(&prize).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = ClaimPrizeResponse{
		PrizeID: prize.ID,
		Status:  prize.Status,
	}
	return resp, nil
}

// PendingComments 获取待审核评论列表
func (a *Admin) PendingComments(info common.PagerForm) (resp PendingCommentsResponse, err error) {
	var comments []model.Comment
	var total int64

	query := model.DB.Model(&model.Comment{}).Where("status = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("User").Preload("Photo").
		Order("created_at ASC").
		Scopes(model.Paginate(common.PagerForm{Page: info.Page, Limit: info.Limit})).
		Find(&comments).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	items := make([]PendingCommentItem, 0, len(comments))
	for _, cm := range comments {
		items = append(items, PendingCommentItem{
			CommentID:  cm.ID,
			PhotoID:    cm.PhotoID,
			PhotoTitle: cm.Photo.Title,
			User:       UserBrief{ID: cm.User.ID, Name: cm.User.Name},
			Comment:    cm.Comment,
			CreatedAt:  cm.CreatedAt,
		})
	}

	resp = PendingCommentsResponse{
		Total: total,
		Items: items,
	}
	return resp, nil
}

// ReviewComment 审核评论
func (a *Admin) ReviewComment(info ReviewCommentParams) (resp ReviewCommentResponse, err error) {
	var comment model.Comment
	if err := model.DB.First(&comment, info.CommentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("评论不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if comment.Status != "pending" {
		return resp, common.ErrNew(errors.New("该评论已审核过"), common.OpErr)
	}

	now := time.Now()

	switch info.Action {
	case "approve":
		comment.Status = "approved"
		comment.ReviewedAt = &now
	case "reject":
		if info.RejectReason == "" {
			return resp, common.ErrNew(errors.New("拒绝时必须填写拒绝原因"), common.ParamErr)
		}
		comment.Status = "rejected"
		comment.RejectReason = info.RejectReason
		comment.ReviewedAt = &now
	default:
		return resp, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	if err := model.DB.Save(&comment).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	msg := "评论已通过审核"
	if info.Action == "reject" {
		msg = "评论已拒绝: " + info.RejectReason
	}

	resp = ReviewCommentResponse{
		CommentID: comment.ID,
		Status:    comment.Status,
		Message:   msg,
	}
	return resp, nil
}
