package service

import (
	"errors"
	"fmt"
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
		query = query.Where("status = ?", info.Status)
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
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var photo model.Photo
	if err := tx.First(&photo, info.PhotoID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if photo.Status != "pending" && info.AdminLevel < 2 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该图片已审核过,请联系更高级管理员修改"), common.OpErr)
	}

	switch info.Action {
	case "approve":
		photo.Status = "approved"
	case "reject":
		if info.RejectReason == "" {
			tx.Rollback()
			return resp, common.ErrNew(errors.New("拒绝时请填写拒绝原因"), common.ParamErr)
		}
		photo.Status = "rejected"
		photo.RejectReason = info.RejectReason
	default:
		tx.Rollback()
		return resp, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	now := time.Now()
	photo.ReviewedAt = &now

	if err := tx.Save(&photo).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	// 发送审核结果消息给投稿用户
	msgSvc := MessageSvc{}
	if err = msgSvc.SendReviewMessage(photo.UserID, info.Action, photo.ID, "photo", info.RejectReason); err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.ID = photo.ID
	resp.Status = photo.Status
	switch info.Action {
	case "approve":
		resp.Message = "图片已通过审核，现已公开"
	case "reject":
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
		query = query.Where("status = ?", info.Status)
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
			LocationSecret:  at.Photo.LocationSecret, //管理员可见
			ImageURL:        at.ImageURL,
			GuessedLocation: at.GuessedLocation,
			Solved:          at.Solved, //管理员可见
			SubmittedAt:     at.CreatedAt,
		})
	}
	return resp, nil
}

// ReviewAttempt 审核答题记录
func (a *Admin) ReviewAttempt(info ReviewAttemptParams) (resp ReviewAttemptResponse, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var attempt model.Attempt
	if err := tx.Preload("Photo").First(&attempt, info.AttemptID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("答题记录不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if attempt.Status != "pending" && info.AdminLevel < 2 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该答题记录已审核过,请联系更高级管理员修改"), common.OpErr)
	}

	now := time.Now()

	switch info.Action {
	case "approve":
		attempt.Status = "approved"
		attempt.ReviewedAt = &now
		//要求管理员等级>2才能审核是否答对，如果审核通过且管理员标记该题为已破解，后续会在事务中处理图片状态和奖品发放
		if info.Solved == "solved" && info.AdminLevel >= 2 { //attempt.GuessedLocation == attempt.Photo.LocationSecret//地址字符完全匹配，若没有审核，则不予通过
			// 判断该题目是否已被破解
			if attempt.Photo.Solved {
				// 已被破解 → 标记为通过但不获奖
				attempt.Solved = 1
			} else {
				// 尚未被破解 → 标记获奖，并更新图片状态
				attempt.Solved = 2 // 2表示管理员审核认为答题正确且给予奖励
				if err := tx.Model(&model.Photo{}).Where("id = ?", attempt.PhotoID).Update("solved", true).Error; err != nil {
					tx.Rollback()
					return resp, common.ErrNew(err, common.SysErr)
				}

				// 生成奖品记录
				prize := &model.Prize{
					UserID:    attempt.UserID,
					PhotoID:   attempt.PhotoID,
					PrizeType: "明信片套装",
					Status:    "unclaimed",
					AwardedAt: &now,
				}
				if err := tx.Create(prize).Error; err != nil {
					tx.Rollback()
					return resp, common.ErrNew(err, common.SysErr)
				}

				// 更新用户获奖次数
				if err := tx.Model(&model.User{}).Where("id = ?", attempt.UserID).
					UpdateColumn("prize_count", gorm.Expr("prize_count + 1")).Error; err != nil {
					tx.Rollback()
					return resp, common.ErrNew(err, common.SysErr)
				}
			}
		} else {
			attempt.Solved = 0 // 标记为不获奖
		}

	case "reject":
		if info.RejectReason == "" {
			tx.Rollback()
			return resp, common.ErrNew(errors.New("拒绝时请填写拒绝原因"), common.ParamErr)
		}
		attempt.Status = "rejected"
		attempt.RejectReason = info.RejectReason
		attempt.ReviewedAt = &now

	default:
		tx.Rollback()
		return resp, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	if err := tx.Save(&attempt).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	// 发送审核结果消息给答题用户
	// if err = msgSvc.SendReviewMessage(attempt.UserID, info.Action, attempt.ID, "attempt", info.RejectReason); err != nil {
	// 	return resp, common.ErrNew(err, common.SysErr)
	// }
	{
		var msgType, title, content string
		if info.Action == "reject" {
			msgType = "review_rejected"
			title = "您的答题未通过审核"
			content = fmt.Sprintf("您提交的答题未通过审核。拒绝原因：%s", info.RejectReason)
		} else if info.Action == "approve" && info.Solved == "unsolved" {
			msgType = "review_approved"
			title = "您的图片已通过审核"
			content = "恭喜！您提交的图片已通过审核；但很遗憾，管理员认为您的答题不正确，未能获得奖品。别气馁，失败是常态，调整心态再试试吧！"
		} else if info.Action == "approve" && info.Solved == "solved" && info.AdminLevel >= 2 && !attempt.Photo.Solved {
			msgType = "review_approved"
			title = "您的图片已通过审核"
			content = "恭喜！您提交的图片已通过审核；您的答题正确，恭喜您获得奖品！"
		} else if info.Action == "approve" && info.Solved == "solved" && attempt.Photo.Solved {
			msgType = "review_approved"
			title = "您的图片已通过审核"
			content = "恭喜！您提交的图片已通过审核；管理员认为您的答题正确，但很遗憾，您的图片已被其他人破解过了，未能获得奖品。别气馁，换一个图片再试试吧！"
		} else {
			msgType = "review_approved"
			title = "您的图片已通过审核"
			content = "恭喜！您提交的图片已通过审核；请等待高级管理员审核答题结果，祝您好运！"
		}

		msg := &model.Message{
			UserID:      attempt.UserID,
			SenderID:    1, // 系统消息
			Type:        msgType,
			Title:       title,
			Content:     content,
			RelatedID:   attempt.ID,
			RelatedType: "attempt",
			IsRead:      false,
		}

		if err := model.DB.Create(msg).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}
	// 审核通过时通知图片作者（自己挑战自己不发）

	msgSvc := MessageSvc{}
	if info.Action == "approve" && attempt.UserID != attempt.Photo.UserID {
		msgSvc.SendAttemptNotification(attempt.UserID, attempt.PhotoID, attempt.Photo.UserID)
	}

	msg := "答题已通过审核"
	if info.Action == "reject" {
		msg = "答题已拒绝: " + info.RejectReason
	}

	resp = ReviewAttemptResponse{
		AttemptID:   attempt.ID,
		Status:      attempt.Status,
		Solved:      attempt.Solved,
		PhotoSolved: attempt.Photo.Solved || attempt.Solved == 2,
		Message:     msg,
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
		Scopes(model.Paginate(info)).
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
			Comment:    cm.CommentText,
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
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var comment model.Comment
	if err := tx.First(&comment, info.CommentID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("评论不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if comment.Status != "pending" {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该评论已审核过"), common.OpErr)
	}

	now := time.Now()

	switch info.Action {
	case "approve":
		comment.Status = "approved"
		comment.ReviewedAt = &now
	case "reject":
		if info.RejectReason == "" {
			tx.Rollback()
			return resp, common.ErrNew(errors.New("拒绝时必须填写拒绝原因"), common.ParamErr)
		}
		comment.Status = "rejected"
		comment.RejectReason = info.RejectReason
		comment.ReviewedAt = &now
	default:
		tx.Rollback()
		return resp, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	if err := tx.Save(&comment).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
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

// ClaimPrize 标记奖品已发放
func (a *Admin) ClaimPrize(prizeID int64) (resp ClaimPrizeResponse, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var prize model.Prize
	if err := tx.First(&prize, prizeID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("奖品记录不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if prize.Status == "claimed" {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该奖品已发放"), common.OpErr)
	}

	prize.Status = "claimed"
	if err := tx.Save(&prize).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	resp = ClaimPrizeResponse{
		PrizeID: prize.ID,
		Status:  prize.Status,
	}
	return resp, nil
}

// UpdateAdminLevel 高级管理员调整其他管理员等级（不超过自身等级）
func (a *Admin) UpdateAdminLevel(info UpdateAdminLevelParams) (resp UpdateAdminLevelResponse, err error) {
	// ----------------仅 Level >= 2 可操作调整管理员等级----------------
	if info.OperatorLevel < 2 {
		return resp, common.ErrNew(errors.New("仅高级管理员可调整管理员等级"), common.LevelErr)
	}
	// 目标等级不能超过操作者自身等级
	if info.TargetLevel > info.OperatorLevel {
		return resp, common.ErrNew(errors.New("目标等级不能超过您自身的等级"), common.LevelErr)
	}

	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var target model.User
	if err := tx.First(&target, info.UserID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("用户不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 不能操作同等级或更高级的管理员
	if target.Level >= info.OperatorLevel {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("无法调整同等级或更高级管理员"), common.LevelErr)
	}

	oldLevel := target.Level
	if oldLevel == info.TargetLevel {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("目标等级与当前等级相同"), common.OpErr)
	}
	// 更新管理员等级(可升可降，但不能超过操作者等级)
	if err := tx.Model(&target).Update("level", info.TargetLevel).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	action := "升级"
	if info.TargetLevel < oldLevel {
		action = "降级"
	}

	resp = UpdateAdminLevelResponse{
		UserID:   target.ID,
		Name:     target.Name,
		OldLevel: oldLevel,
		NewLevel: info.TargetLevel,
		Message:  "管理员" + target.Name + fmt.Sprintf("( %d )", target.ID) + action + "成功",
	}
	return resp, nil
}
