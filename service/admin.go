package service

import (
	"errors"
	"fmt"
	"sort"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type AdminSvc struct{}

// PendingPhotos 获取待审核图片列表
func (a *AdminSvc) PendingPhotos(info AdminPendingPhotoParams) (resp AdminPendingPhotoForms, err error) {
	var photos []model.Photo
	var total int64
	query := model.DB.Model(&model.Photo{})
	if info.AdminLevel < 3 {
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
	resp.PendingPhotos = make([]AdminPendingPhotoForm, 0, len(photos))
	for _, photo := range photos {
		resp.PendingPhotos = append(resp.PendingPhotos, AdminPendingPhotoForm{
			ID:          photo.ID,
			UserID:      photo.UserID,
			ActivityID:  photo.ActivityID,
			Title:       photo.Title,
			Description: photo.Description,
			Longitude:   photo.Longitude,
			Latitude:    photo.Latitude,
			ThumbURL:    photo.ThumbURL,
		})
	}
	return resp, nil
}

// ReviewPhoto 审核图片
func (a *AdminSvc) ReviewPhoto(info AdminReviewPhotoParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var photo model.Photo
	var msgType, title, content string
	if err := tx.Preload("Activity").First(&photo, info.PhotoID).Error; err != nil {
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
		msgType = "review_approved"
		title = "您的图片投稿已通过审核"
		content = "恭喜！您提交的图片投稿已通过审核。"
		scoreParams := ScoreChangeParams{
			UserID:      photo.UserID,
			Delta:       photo.Activity.PhotoPoints,
			Reason:      "upload_photo",
			RelatedID:   photo.ID,
			RelatedType: "photo",
		}
		scoreSvc := ScoreSvc{}
		if _, err := scoreSvc.RegularScoreChange(scoreParams); err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
	case "reject":
		if info.RejectReason == "" {
			tx.Rollback()
			return resp, common.ErrNew(errors.New("拒绝时请填写拒绝原因"), common.ParamErr)
		}
		photo.Status = "rejected"
		photo.RejectReason = info.RejectReason
		msgType = "review_rejected"
		title = "您的图片投稿未通过审核"
		content = "您提交的图片投稿未通过审核。拒绝原因：" + info.RejectReason
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

	// 奖励积分（审核通过时）
	if info.Action == "approve" {

	}

	// 发送审核结果消息给投稿用户
	msg := &model.Message{
		UserID:      photo.UserID,
		SenderID:    1, // 系统消息
		Type:        msgType,
		Title:       title,
		Content:     content,
		RelatedID:   info.PhotoID,
		RelatedType: "photo",
		IsRead:      false,
	}

	if err := tx.Create(msg).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	resp.ID = photo.ID
	resp.Status = photo.Status
	return resp, nil
}

// PendingAttempts 获取待审核答题记录
func (a *AdminSvc) PendingAttempts(info AdminPendingAttemptParams) (resp AdminPendingAttemptForms, err error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{})

	if info.AdminLevel < 3 {
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
	resp.Attempts = make([]AdminPendingAttemptForm, 0, len(attempts))
	for _, at := range attempts {
		resp.Attempts = append(resp.Attempts, AdminPendingAttemptForm{
			AttemptID:      at.ID,
			PhotoID:        at.PhotoID,
			PhotoTitle:     at.Photo.Title,
			GuassThumbURL:  at.ImageURL,
			GuassLongitude: at.Longitude,
			GuassLatitude:  at.Latitude,
			ThumbURL:       at.Photo.ThumbURL,
			Longitude:      at.Photo.Longitude,
			Latitude:       at.Photo.Latitude,
			Status:         at.Status, // 管理员可见
			SubmittedAt:    at.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}

// ReviewAttempt 审核答题记录
func (a *AdminSvc) ReviewAttempt(info AdminReviewAttemptParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var attempt model.Attempt
	var msgType, title, content string
	var delta, awardedBatch int

	if err := tx.Preload("Photo.Activity.AttemptRewardTiers").First(&attempt, info.AttemptID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("答题记录不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if attempt.Status != "pending" && info.AdminLevel < 3 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该答题记录已审核过,请联系更高级管理员修改"), common.OpErr)
	}

	now := time.Now()

	switch info.Solved {
	case "solved":
		// 奖励积分（审核通过时）（自己挑战自己不发）
		if attempt.UserID != attempt.Photo.UserID {
			activitySvc := ActivitySvc{}
			rank, err := activitySvc.GetUserRank(attempt.UserID, attempt.PhotoID)
			if err != nil {
				tx.Rollback()
				return resp, common.ErrNew(err, common.SysErr)
			}
			// 1. 获取奖励配置切片
			tiers := attempt.Photo.Activity.AttemptRewardTiers

			// 2. 按批次从小到大排序（确保先匹配最严格的区间）
			sort.Slice(tiers, func(i, j int) bool {
				return tiers[i].Batch < tiers[j].Batch
			})

			// 3. 计算奖励积分和批次
			if rank > 0 { // rank=0 表示未答对或未上榜
				for _, tier := range tiers {
					if rank <= tier.RankLimit {
						delta = tier.AttemptPoints
						awardedBatch = tier.Batch
						break
					}
				}
			}
			scoreParams := ScoreChangeParams{
				UserID:      attempt.UserID,
				Delta:       delta,
				Reason:      "upload_photo",
				RelatedID:   attempt.ID,
				RelatedType: "attempt",
				Remark:      fmt.Sprintf("恭喜你答对了，是第 %d 批次，得分 %d ！", awardedBatch, delta),
			}
			scoreSvc := ScoreSvc{}
			if _, err := scoreSvc.RegularScoreChange(scoreParams); err != nil {
				tx.Rollback()
				return resp, common.ErrNew(err, common.SysErr)
			}
		}

		attempt.Status = "solved"
		attempt.ReviewedAt = &now
		msgType = "review_approved"
		title = "您的答题正确"
		content = fmt.Sprintf("恭喜你答对了，将得到积分 %d ！", delta)

		if err := tx.Model(&model.Photo{}).Where("id = ?", attempt.PhotoID).Update("solved", true).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
	case "unsolved":
		attempt.Status = "unsolved"
		attempt.ReviewedAt = &now
		msgType = "review_rejected"
		title = "您的答题不正确"
		if info.RejectReason != "" {
			content = "您提交的答题不正确。拒绝原因：" + info.RejectReason
			attempt.RejectReason = info.RejectReason
		} else {
			attempt.RejectReason = "您的答题不正确，未能获得奖品。别气馁，失败是常态，调整心态再试试吧！"
			content = "您的答题不正确，未能获得奖品。别气馁，失败是常态，调整心态再试试吧！"
		}
	default:
		tx.Rollback()
		return resp, common.ErrNew(errors.New("action 必须为 approve 或 reject"), common.ParamErr)
	}

	if err := tx.Save(&attempt).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
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

	if err := tx.Create(msg).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     attempt.ID,
		Status: attempt.Status,
	}
	return resp, nil
}

// PendingComments 获取待审核评论列表
func (a *AdminSvc) PendingComments(info AdminPendingCommentParams) (resp AdminPendingCommentForms, err error) {
	var comments []model.Comment
	var total int64

	query := model.DB.Model(&model.Comment{})
	if info.AdminLevel < 3 {
		query = query.Where("status = ?", "pending")
	} else {
		query = query.Where("status = ?", info.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("User").Preload("Photo").
		Order("created_at ASC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&comments).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	items := make([]AdminPendingCommentForm, 0, len(comments))
	for _, cm := range comments {
		items = append(items, AdminPendingCommentForm{
			CommentID:  cm.ID,
			PhotoID:    cm.PhotoID,
			PhotoTitle: cm.Photo.Title,
			User:       UserBrief{ID: cm.User.ID, Nickname: cm.User.Nickname},
			Comment:    cm.CommentText,
			CreatedAt:  cm.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	resp = AdminPendingCommentForms{
		Total: total,
		Items: items,
	}
	return resp, nil
}

// ReviewComment 审核评论
func (a *AdminSvc) ReviewComment(info AdminReviewCommentParams) (resp ResponseIS, err error) {
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
		msg := &model.Message{
			UserID:      comment.UserID,
			SenderID:    1, // 系统消息
			Type:        "review_rejected",
			Title:       "您的评论审核未通过",
			Content:     "您的评论审核未通过，拒绝原因：" + info.RejectReason,
			RelatedID:   comment.ID,
			RelatedType: "comment",
			IsRead:      false,
		}

		if err := tx.Create(msg).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
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

	resp = ResponseIS{
		ID:     comment.ID,
		Status: comment.Status,
	}
	return resp, nil
}

// Announcement 全服公告
func (a *AdminSvc) Announcement(info AdminAnnouncement) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	msg := &model.Message{
		UserID:      -1, //全体用户
		SenderID:    1,  // 系统消息
		Type:        "notice",
		Title:       info.Title,
		Content:     info.Content,
		RelatedID:   info.RelatedID,
		RelatedType: info.RelatedType,
		IsRead:      false,
	}

	if err := tx.Create(msg).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp = ResponseIS{
		ID:     msg.ID,
		Status: "success",
	}
	return resp, nil
}

// ListGoods 管理员获取所有奖品列表（已分发/未分发）
func (a *AdminSvc) AdminGoodList(info AdminListGoodsParams) (resp AdminGoodForms, err error) {
	var goods []model.Good
	var total int64

	query := model.DB.Model(&model.Good{})

	if info.Available {
		query = query.Where(gorm.Expr("stock > ?", 0))
	}
	if info.Status != "" {
		query = query.Where("status = ?", info.Status)
	}
	if info.Keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+info.Keyword+"%", "%"+info.Keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("awarded_at DESC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&goods).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Goods = make([]AdminGoodForm, 0, len(goods))
	for _, g := range goods {
		resp.Goods = append(resp.Goods, AdminGoodForm{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			ThumbURL:    g.ThumbURL,
			NeedScore:   g.NeedScore,
			Stock:       g.Stock,
			Status:      g.Status,
			CreatedAt:   g.BaseModel.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return resp, nil
}

// AdminGetByID 获取奖品详情
func (a *AdminSvc) AdminGetByID(params AdminGoodGetByIDParams) (resp AdminGoodDetail, err error) {
	var good model.Good
	if err := model.DB.First(&good, params.GoodID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, errors.New("奖品不存在")
		}
		return resp, err
	}

	resp = AdminGoodDetail{
		ID:          good.ID,
		Name:        good.Name,
		Description: good.Description,
		ImageURL:    good.ImageURL,
		NeedScore:   good.NeedScore,
		Stock:       good.Stock,
		Status:      good.Status,
		CreatedAt:   good.BaseModel.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

// // ClaimPrize 标记奖品已发放
// func (a *AdminSvc) ClaimPrize(prizeID int64) (resp AdminClaimPrizeResponse, err error) {
// 	tx := model.DB.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 			panic(r)
// 		}
// 	}()

// 	var prize model.Prize
// 	if err := tx.First(&prize, prizeID).Error; err != nil {
// 		tx.Rollback()
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return resp, common.ErrNew(errors.New("奖品记录不存在"), common.OpErr)
// 		}
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	if prize.Status == "claimed" {
// 		tx.Rollback()
// 		return resp, common.ErrNew(errors.New("该奖品已发放"), common.OpErr)
// 	}

// 	prize.Status = "claimed"
// 	if err := tx.Save(&prize).Error; err != nil {
// 		tx.Rollback()
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
// 	}

// 	resp = AdminClaimPrizeResponse{
// 		PrizeID: prize.ID,
// 		Status:  prize.Status,
// 	}
// 	return resp, nil
// }

// UpdateAdminLevel 高级管理员调整其他管理员等级（不超过自身等级）
func (a *AdminSvc) UpdateAdminLevel(info AdminUpdateLevelParams) (resp ResponseIS, err error) {
	// ----------------仅 Level >= 4 可操作调整管理员等级----------------
	if info.OperatorLevel < 3 {
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

	resp = ResponseIS{
		ID:     target.ID,
		Status: "success",
	}
	return resp, nil
}
