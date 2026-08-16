package service

import (
	"errors"
	"fmt"
	"sort"
	"time"
	"tu-xun/common"
	"tu-xun/model"
	"tu-xun/pkg/urlutil"

	"gorm.io/gorm"
)

type AdminSvc struct{}

// ==================== Photo Management ====================

// CreatePhoto 管理员新增题目（官方账号 user_id=1，直接通过审核）
func (a *AdminSvc) CreatePhoto(form AdminPhotoCreateForm) (resp ResponseIS, err error) {
	if form.Title == "" {
		return resp, common.ErrNew(errors.New("标题不能为空"), common.ParamErr)
	}
	if form.ImageFile == nil {
		return resp, common.ErrNew(errors.New("图片不能为空"), common.ParamErr)
	}
	if form.CoordType != "" && form.CoordType != "wgs84" && form.CoordType != "gcj02" && form.CoordType != "bd09" {
		return resp, common.ErrNew(errors.New("坐标系类型无效"), common.ParamErr)
	}

	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 校验活动存在且未结束
	var activity model.Activity
	if err := tx.Where("id = ?", form.ActivityID).First(&activity).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("活动不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}
	if !activity.IsActive && activity.EndTime != nil && !time.Now().Before(*activity.EndTime) {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("活动已结束，无法新增题目"), common.OpErr)
	}

	// 上传图片
	uploadResult, err := saveUploadedFile(form.ImageFile, "photos", true)
	if err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 坐标转换
	gcjLat, gcjLng := WGS84orGCJ02ToGCJ02(form.Latitude, form.Longitude, form.CoordType)

	photo := &model.Photo{
		UserID:        1, // 官方账号
		ActivityID:    form.ActivityID,
		Title:         form.Title,
		Description:   form.Description,
		Latitude:      gcjLat,
		Longitude:     gcjLng,
		CoordType:     "gcj02",
		ImageURL:      uploadResult.ImageURL,
		ThumbURL:      uploadResult.ThumbURL,
		ImageWidth:    uploadResult.ImageWidth,
		ImageHeight:   uploadResult.ImageHeight,
		ThumbWidth:    uploadResult.ThumbWidth,
		ThumbHeight:   uploadResult.ThumbHeight,
		Status:        "approved",
		Solved:        false,
		SolvedCount:   0,
		AttemptsCount: 0,
		LikesCount:    0,
	}

	if err := tx.Create(photo).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return ResponseIS{ID: photo.ID, Status: photo.Status}, nil
}

// UpdatePhoto 管理员编辑题目内容（不改动审核状态）
func (a *AdminSvc) UpdatePhoto(form AdminPhotoUpdateForm) (resp ResponseIS, err error) {
	// 至少提供一个更新字段
	if form.Title == "" && form.Description == "" && form.ImageFile == nil &&
		form.Longitude == 0 && form.Latitude == 0 && form.CoordType == "" {
		return resp, common.ErrNew(errors.New("至少提供一个更新字段"), common.ParamErr)
	}

	// 坐标字段必须全有或全无
	hasLng := form.Longitude != 0
	hasLat := form.Latitude != 0
	hasCoord := form.CoordType != ""
	if (hasLng || hasLat || hasCoord) && !(hasLng && hasLat && hasCoord) {
		return resp, common.ErrNew(errors.New("经度、纬度和坐标系类型必须同时提供"), common.ParamErr)
	}

	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var photo model.Photo
	if err := tx.Where("id = ?", form.PhotoID).First(&photo).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("题目不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	updates := map[string]interface{}{}

	if form.Title != "" {
		updates["title"] = form.Title
	}
	if form.Description != "" {
		updates["description"] = form.Description
	}
	if form.ImageFile != nil {
		uploadResult, err := saveUploadedFile(form.ImageFile, "photos", true)
		if err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
		updates["image_url"] = uploadResult.ImageURL
		updates["thumb_url"] = uploadResult.ThumbURL
		updates["image_width"] = uploadResult.ImageWidth
		updates["image_height"] = uploadResult.ImageHeight
		updates["thumb_width"] = uploadResult.ThumbWidth
		updates["thumb_height"] = uploadResult.ThumbHeight
	}
	if hasLng && hasLat && hasCoord {
		gcjLat, gcjLng := WGS84orGCJ02ToGCJ02(form.Latitude, form.Longitude, form.CoordType)
		updates["latitude"] = gcjLat
		updates["longitude"] = gcjLng
		updates["coord_type"] = "gcj02"
	}

	if len(updates) == 0 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("没有有效的更新字段"), common.ParamErr)
	}

	if err := tx.Model(&photo).Updates(updates).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return ResponseIS{ID: photo.ID, Status: photo.Status}, nil
}

// ListPhotos 获取题目池列表
func (a *AdminSvc) ListPhotos(params AdminPhotoListParams) (resp AdminPhotoListPage, err error) {
	var photos []model.Photo
	var total int64
	query := model.DB.Model(&model.Photo{}).Preload("Activity").Preload("Author")

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if len(params.ActivityIDs) > 0 {
		query = query.Where("activity_id IN ?", params.ActivityIDs)
	}
	if params.Solved != nil {
		query = query.Where("solved = ?", *params.Solved)
	}
	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", kw, kw)
	}
	if params.UserKeyword != "" {
		ukw := "%" + params.UserKeyword + "%"
		query = query.Where("user_id IN (SELECT id FROM user WHERE nickname LIKE ? OR name LIKE ? OR netid LIKE ?)", ukw, ukw, ukw)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at ASC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&photos).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]AdminPhotoListItem, 0, len(photos))
	for _, photo := range photos {
		resp.List = append(resp.List, AdminPhotoListItem{
			ID:          photo.ID,
			Activity:    ActivityBrief{ID: photo.Activity.ID, Title: photo.Activity.Title},
			Author:      UserBrief{ID: photo.Author.ID, Nickname: photo.Author.Nickname, Avatar: urlutil.FullURL(photo.Author.AvatarURL)},
			Title:       photo.Title,
			Description: photo.Description,
			Image: Media{
				OriginURL: urlutil.FullURL(photo.ImageURL),
				ThumbURL:  urlutil.FullURL(photo.ThumbURL),
				Width:     photo.ImageWidth,
				Height:    photo.ImageHeight,
			},
			Location: Location{
				Longitude: photo.Longitude,
				Latitude:  photo.Latitude,
				CoordType: photo.CoordType,
			},
			AttemptsCount: photo.AttemptsCount,
			SolvedCount:   photo.SolvedCount,
			LikesCount:    photo.LikesCount,
			Status:        photo.Status,
			RejectReason:  photo.RejectReason,
			CreatedAt:     &photo.CreatedAt,
		})
	}
	return resp, nil
}

// ReviewPhoto 审核图片
func (a *AdminSvc) ReviewPhoto(params AdminReviewPhotoParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var photo model.Photo
	if err := tx.Preload("Activity").First(&photo, params.PhotoID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 409: 已审核过的不能再次审核
	if photo.Status == "approved" || photo.Status == "rejected" {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该图片已审核过"), common.ConflictErr)
	}

	var msgType, content string

	switch params.Action {
	case "approve":
		photo.Status = "approved"
		msgType = "review_approved"
		content = "恭喜！您提交的图片投稿已通过审核。"

		scoreParams := ScoreChangeParams{
			UserID:      photo.UserID,
			Delta:       photo.Activity.PhotoPoints,
			Reason:      "upload_photo",
			RelatedID:   photo.ID,
			RelatedType: "photo",
		}
		scoreSvc := ScoreSvc{}
		if _, err := scoreSvc.RegularScoreChange(tx, scoreParams); err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
	case "reject":
		if params.RejectReason == "" {
			tx.Rollback()
			return resp, common.ErrNew(errors.New("拒绝时请填写拒绝原因"), common.ParamErr)
		}
		photo.Status = "rejected"
		photo.RejectReason = params.RejectReason
		msgType = "review_rejected"
		content = "您提交的图片投稿未通过审核。拒绝原因：" + params.RejectReason
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

	// 发送审核结果消息给投稿用户
	msg := &model.InteractionMessage{
		UserID:      photo.UserID,
		SenderID:    1,
		Type:        msgType,
		Content:     content,
		RelatedID:   params.PhotoID,
		RelatedType: "photo",
		PhotoID:     params.PhotoID,
		IsRead:      false,
	}
	if err := tx.Create(msg).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return ResponseIS{ID: photo.ID, Status: photo.Status}, nil
}

// ==================== Attempt Management ====================

// ListAttempts 获取答题列表
func (a *AdminSvc) ListAttempts(params AdminAttemptListParams) (resp AdminAttemptListPage, err error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		query = query.Where("comment_text LIKE ?", kw)
	}
	if params.PhotoKeyword != "" {
		pkw := "%" + params.PhotoKeyword + "%"
		query = query.Where("photo_id IN (SELECT id FROM photo WHERE title LIKE ?)", pkw)
	}
	if params.UserKeyword != "" {
		ukw := "%" + params.UserKeyword + "%"
		query = query.Where("user_id IN (SELECT id FROM user WHERE nickname LIKE ? OR name LIKE ? OR netid LIKE ?)", ukw, ukw, ukw)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("Photo").Preload("User").
		Order("created_at ASC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]AdminAttemptListItem, 0, len(attempts))
	for _, at := range attempts {
		resp.List = append(resp.List, AdminAttemptListItem{
			ID: at.ID,
			User: UserBrief{
				ID:       at.User.ID,
				Nickname: at.User.Nickname,
				Avatar:   urlutil.FullURL(at.User.AvatarURL),
			},
			Photo: AdminAttemptPhotoBrief{
				ID:    at.Photo.ID,
				Title: at.Photo.Title,
				Image: Media{
					ThumbURL: urlutil.FullURL(at.Photo.ThumbURL),
					Width:    at.Photo.ThumbWidth,
					Height:   at.Photo.ThumbHeight,
				},
				Location: Location{
					Longitude: at.Photo.Longitude,
					Latitude:  at.Photo.Latitude,
					CoordType: at.Photo.CoordType,
				},
			},
			GuessImage: Media{
				ThumbURL: urlutil.FullURL(at.ImageURL),
				Width:    at.ImageWidth,
				Height:   at.ImageHeight,
			},
			GuessLocation: Location{
				Longitude: at.Longitude,
				Latitude:  at.Latitude,
				CoordType: at.CoordType,
			},
			Status:       at.Status,
			RejectReason: at.RejectReason,
			CreatedAt:    &at.CreatedAt,
		})
	}
	return resp, nil
}

// ReviewAttempt 审核答题记录
func (a *AdminSvc) ReviewAttempt(params AdminReviewAttemptParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var attempt model.Attempt
	if err := tx.Preload("Photo.Activity.AttemptRewardTiers").First(&attempt, params.AttemptID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("答题记录不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 409: 已审核过的不能再次审核
	if attempt.Status == "solved" || attempt.Status == "unsolved" {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该答题记录已审核过"), common.ConflictErr)
	}

	var msgType, content string
	now := time.Now()

	switch params.Solved {
	case "solved":
		// 先更新状态到事务，以便 GetUserRank 能查询到
		attempt.Status = "solved"
		attempt.ReviewedAt = &now
		if err := tx.Model(&attempt).Updates(map[string]any{
			"status":      "solved",
			"reviewed_at": now,
		}).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}

		// 奖励积分（自己挑战自己不发）
		if attempt.UserID != attempt.Photo.UserID {
			activitySvc := ActivitySvc{}
			rank, err := activitySvc.GetUserRank(tx, attempt.UserID, attempt.PhotoID)
			if err != nil {
				tx.Rollback()
				return resp, common.ErrNew(err, common.SysErr)
			}

			var delta, awardedBatch int
			tiers := attempt.Photo.Activity.AttemptRewardTiers
			sort.Slice(tiers, func(i, j int) bool {
				return tiers[i].Batch < tiers[j].Batch
			})

			if rank > 0 {
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
			if _, err := scoreSvc.RegularScoreChange(tx, scoreParams); err != nil {
				tx.Rollback()
				return resp, common.ErrNew(err, common.SysErr)
			}
		}

		msgType = "review_approved"
		content = "恭喜！您的答题已通过审核，判定为正确。"

		// 递增 photo.solved_count
		if err := tx.Model(&model.Photo{}).Where("id = ?", attempt.PhotoID).
			Update("solved_count", gorm.Expr("solved_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}

	case "unsolved":
		attempt.Status = "unsolved"
		attempt.ReviewedAt = &now
		attempt.RejectReason = params.RejectReason
		msgType = "review_rejected"
		if params.RejectReason != "" {
			content = "您的答题未通过审核。拒绝原因：" + params.RejectReason
		} else {
			content = "您的答题未通过审核。别气馁，失败是常态，调整心态再试试吧！"
		}
	default:
		tx.Rollback()
		return resp, common.ErrNew(errors.New("solved 必须为 solved 或 unsolved"), common.ParamErr)
	}

	if err := tx.Save(&attempt).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	msg := &model.InteractionMessage{
		UserID:      attempt.UserID,
		SenderID:    1,
		Type:        msgType,
		Content:     content,
		RelatedID:   attempt.ID,
		RelatedType: "attempt",
		PhotoID:     attempt.PhotoID,
		IsRead:      false,
	}
	if err := tx.Create(msg).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return ResponseIS{ID: attempt.ID, Status: attempt.Status}, nil
}

// ==================== Comment Management ====================

// ListComments 获取评论列表
func (a *AdminSvc) ListComments(params AdminCommentListParams) (resp AdminCommentListPage, err error) {
	var comments []model.Comment
	var total int64

	query := model.DB.Model(&model.Comment{})
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		query = query.Where("comment_text LIKE ?", kw)
	}
	if params.PhotoKeyword != "" {
		pkw := "%" + params.PhotoKeyword + "%"
		query = query.Where("photo_id IN (SELECT id FROM photo WHERE title LIKE ?)", pkw)
	}
	if params.UserKeyword != "" {
		ukw := "%" + params.UserKeyword + "%"
		query = query.Where("user_id IN (SELECT id FROM user WHERE nickname LIKE ? OR name LIKE ? OR netid LIKE ?)", ukw, ukw, ukw)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("User").Preload("Photo").
		Order("created_at ASC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&comments).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]AdminCommentListItem, 0, len(comments))
	for _, cm := range comments {
		resp.List = append(resp.List, AdminCommentListItem{
			ID: cm.ID,
			User: UserBrief{
				ID:       cm.User.ID,
				Nickname: cm.User.Nickname,
				Avatar:   urlutil.FullURL(cm.User.AvatarURL),
			},
			Photo: AdminCommentPhotoBrief{
				ID:    cm.Photo.ID,
				Title: cm.Photo.Title,
			},
			Content:   cm.CommentText,
			Status:    cm.Status,
			CreatedAt: &cm.CreatedAt,
		})
	}
	return resp, nil
}

// ReviewComment 审核评论（可在 approved 和 rejected 之间切换，无 409）
func (a *AdminSvc) ReviewComment(params AdminReviewCommentParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var comment model.Comment
	if err := tx.First(&comment, params.CommentID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("评论不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	now := time.Now()

	switch params.Action {
	case "approve":
		comment.Status = "approved"
		comment.ReviewedAt = &now
	case "reject":
		if params.RejectReason == "" {
			tx.Rollback()
			return resp, common.ErrNew(errors.New("拒绝时必须填写拒绝原因"), common.ParamErr)
		}
		comment.Status = "rejected"
		comment.RejectReason = params.RejectReason
		comment.ReviewedAt = &now

		msg := &model.InteractionMessage{
			UserID:      comment.UserID,
			SenderID:    1,
			Type:        "review_rejected",
			Content:     "您的评论审核未通过，拒绝原因：" + params.RejectReason,
			RelatedID:   comment.ID,
			RelatedType: "comment",
			PhotoID:     comment.PhotoID,
			IsRead:      false,
		}
		if err := tx.Create(msg).Error; err != nil {
			tx.Rollback()
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

	return ResponseIS{ID: comment.ID, Status: comment.Status}, nil
}

// ==================== User Management (Level 3 only) ====================

// ListUsers 关键词搜索用户列表（Level 3 专用）
func (a *AdminSvc) ListUsers(params AdminUserListParams) (resp AdminUserPage, err error) {
	var users []model.User
	var total int64

	query := model.DB.Model(&model.User{})

	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		query = query.Where("id LIKE ? OR netid LIKE ? OR name LIKE ? OR nickname LIKE ?", kw, kw, kw, kw)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Level != 0 {
		query = query.Where("level = ?", params.Level)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("id ASC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&users).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]UserSummary, 0, len(users))
	for _, u := range users {
		resp.List = append(resp.List, UserSummary{
			ID:                     u.ID,
			NetID:                  u.NetID,
			Username:               u.Name,
			Nickname:               u.Nickname,
			Avatar:                 urlutil.FullURL(u.AvatarURL),
			Level:                  u.Level,
			ScoreCount:             u.ScoreCount,
			Status:                 u.Status,
			NicknameEditsRemaining: 0,
			AvatarEditsRemaining:   0,
		})
	}
	return resp, nil
}

// UpdateAdminLevel 高级管理员调整其他管理员等级（Level 3 专用）
func (a *AdminSvc) UpdateAdminLevel(params AdminUpdateLevelParams) error {
	// Level 3 校验
	if params.OperatorLevel < 3 {
		return common.ErrNew(errors.New("仅高级管理员可调整管理员等级"), common.LevelErr)
	}
	// 目标等级不能超过操作者自身等级
	if params.TargetLevel > params.OperatorLevel {
		return common.ErrNew(errors.New("目标等级不能超过您自身的等级"), common.LevelErr)
	}

	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var targetUser model.User
	if err := tx.First(&targetUser, params.UserID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNew(errors.New("用户不存在"), common.OpErr)
		}
		return common.ErrNew(err, common.SysErr)
	}

	// 不能操作自己
	if targetUser.ID == params.OperatorID {
		tx.Rollback()
		return common.ErrNew(errors.New("不能操作自己"), common.ParamErr)
	}

	// 不能操作 Level 3 账户
	if targetUser.Level >= 3 {
		tx.Rollback()
		return common.ErrNew(errors.New("无法操作高级管理员"), common.LevelErr)
	}

	// 不能操作同等级或更高级管理员
	if targetUser.Level >= params.OperatorLevel {
		tx.Rollback()
		return common.ErrNew(errors.New("无法调整同等级或更高级管理员"), common.LevelErr)
	}

	if err := tx.Model(&targetUser).Update("level", params.TargetLevel).Error; err != nil {
		tx.Rollback()
		return common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return nil
}

// SetUserStatus 封禁/解封用户（Level 3 专用）
func (a *AdminSvc) SetUserStatus(params AdminSetUserStatusParams) error {
	// Level 3 校验
	if params.OperatorLevel < 3 {
		return common.ErrNew(errors.New("仅高级管理员可封禁/解封用户"), common.LevelErr)
	}

	var targetUser model.User
	if err := model.DB.First(&targetUser, params.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNew(errors.New("用户不存在"), common.OpErr)
		}
		return common.ErrNew(err, common.SysErr)
	}

	// 不能操作自己
	if targetUser.ID == params.OperatorID {
		return common.ErrNew(errors.New("不能封禁/解封自己"), common.ParamErr)
	}

	// 不能操作 Level 3 账户
	if targetUser.Level >= 3 {
		return common.ErrNew(errors.New("无法操作高级管理员"), common.LevelErr)
	}

	// 幂等：相同状态直接返回
	if targetUser.Status == params.Status {
		return nil
	}

	if err := model.DB.Model(&targetUser).Update("status", params.Status).Error; err != nil {
		return common.ErrNew(err, common.SysErr)
	}

	return nil
}
