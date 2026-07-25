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

// ListPhotos 获取题目池列表
func (a *AdminSvc) ListPhotos(info AdminPhotoListParams) (resp AdminPhotoListForms, err error) {
	var photos []model.Photo
	var total int64
	query := model.DB.Model(&model.Photo{}).Preload("Activity").Preload("Author")

	if info.Status != "" {
		query = query.Where("status = ?", info.Status)
	}
	if len(info.ActivityIDs) > 0 {
		query = query.Where("activity_id IN ?", info.ActivityIDs)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", kw, kw)
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
	resp.List = make([]AdminPhotoListItem, 0, len(photos))
	for _, photo := range photos {
		resp.List = append(resp.List, AdminPhotoListItem{
			Activity:      ActivityBrief{ID: photo.Activity.ID, Title: photo.Activity.Title},
			Author:        UserBrief{ID: photo.Author.ID, Nickname: photo.Author.Nickname, AvatarURL: photo.Author.AvatarURL},
			ID:            photo.ID,
			Title:         photo.Title,
			Description:   photo.Description,
			ImageURL:      photo.ImageURL,
			ThumbURL:      photo.ThumbURL,
			Longitude:     photo.Longitude,
			Latitude:      photo.Latitude,
			CoordType:     "",
			Solved:        photo.Solved,
			AttemptsCount: photo.AttemptsCount,
			LikesCount:    photo.LikesCount,
			Status:        photo.Status,
			RejectReason:  photo.RejectReason,
			CreatedAt:     &photo.CreatedAt,
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

	if photo.Status != "pending" && info.AdminLevel < 3 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该图片已审核过,请联系更高级管理员修改"), common.ConflictErr)
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
		if _, err := scoreSvc.RegularScoreChange(tx, scoreParams); err != nil {
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
			Category:    "normal",
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

// CreatePhoto 管理员新增题目（使用官方账号 user_id=1，直接通过审核）
func (a *AdminSvc) CreatePhoto(activityID int64, form AdminPhotoUpsertForm) (resp ResponseIS, err error) {
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

	// 校验活动存在
	var activity model.Activity
	if err := tx.Where("id = ?", activityID).First(&activity).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("活动不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 上传图片
	imageURL, thumbURL, err := saveUploadedFile(form.ImageFile, "photos", true)
	if err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 坐标转换
	gcjLat, gcjLng := WGS84orGCJ02ToGCJ02(form.Latitude, form.Longitude, form.CoordType)

	photo := &model.Photo{
		UserID:        1, // 官方账号
		ActivityID:    activityID,
		Title:         form.Title,
		Description:   form.Description,
		Latitude:      gcjLat,
		Longitude:     gcjLng,
		ImageURL:      imageURL,
		ThumbURL:      thumbURL,
		Status:        "approved",
		Solved:        false,
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

	resp = ResponseIS{
		ID:     photo.ID,
		Status: photo.Status,
	}
	return resp, nil
}

// UpdatePhoto 管理员编辑题目内容（不改动审核状态）
func (a *AdminSvc) UpdatePhoto(activityID, photoID int64, form AdminPhotoUpsertForm) (resp ResponseIS, err error) {
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
	if err := tx.Where("id = ? AND activity_id = ?", photoID, activityID).First(&photo).Error; err != nil {
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
		imageURL, thumbURL, err := saveUploadedFile(form.ImageFile, "photos", true)
		if err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
		updates["image_url"] = imageURL
		updates["thumb_url"] = thumbURL
	}
	if hasLng && hasLat && hasCoord {
		gcjLat, gcjLng := WGS84orGCJ02ToGCJ02(form.Latitude, form.Longitude, form.CoordType)
		updates["latitude"] = gcjLat
		updates["longitude"] = gcjLng
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

	resp = ResponseIS{
		ID:     photo.ID,
		Status: photo.Status,
	}
	return resp, nil
}

// ListAttempts 获取答题列表
func (a *AdminSvc) ListAttempts(info AdminAttemptListParams) (resp AdminAttemptListForms, err error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{})

	if info.Status != "" {
		query = query.Where("status = ?", info.Status)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		query = query.Where("comment_text LIKE ?", kw)
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
	resp.List = make([]AdminAttemptListItem, 0, len(attempts))
	for _, at := range attempts {
		resp.List = append(resp.List, AdminAttemptListItem{
			AttemptID:      at.ID,
			PhotoID:        at.PhotoID,
			PhotoTitle:     at.Photo.Title,
			GuessImageURL:  at.ImageURL,
			GuessLongitude: at.Longitude,
			GuessLatitude:  at.Latitude,
			GuessCoordType: "",
			ThumbURL:       at.Photo.ThumbURL,
			Longitude:      at.Photo.Longitude,
			Latitude:       at.Photo.Latitude,
			CoordType:      "",
			Status:         at.Status,
			SubmittedAt:    &at.CreatedAt,
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
			if _, err := scoreSvc.RegularScoreChange(tx, scoreParams); err != nil {
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

	// 更新投稿答题次数
	result := tx.Model(&model.Photo{}).
		Where("id = ?", attempt.PhotoID).
		Update("attempts_count", gorm.Expr("attempts_count + ?", 1))

	// 处理更新错误
	if result.Error != nil {
		tx.Rollback()
		return resp, common.ErrNew(result.Error, common.SysErr)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("没有找到对应的图片投稿"), common.SysErr)
	}

	msg := &model.Message{
		UserID:      attempt.UserID,
		SenderID:    1, // 系统消息
			Category:    "normal",
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

// ListComments 获取评论列表
func (a *AdminSvc) ListComments(info AdminCommentListParams) (resp AdminCommentListForms, err error) {
	var comments []model.Comment
	var total int64

	query := model.DB.Model(&model.Comment{})
	if info.Status != "" {
		query = query.Where("status = ?", info.Status)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		query = query.Where("comment_text LIKE ?", kw)
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

	items := make([]AdminCommentListItem, 0, len(comments))
	for _, cm := range comments {
		items = append(items, AdminCommentListItem{
			CommentID:  cm.ID,
			PhotoID:    cm.PhotoID,
			PhotoTitle: cm.Photo.Title,
			User:       UserBrief{ID: cm.User.ID, Nickname: cm.User.Nickname, AvatarURL: cm.User.AvatarURL},
			Comment:    cm.CommentText,
			Status:     cm.Status,
			CreatedAt:  &cm.CreatedAt,
		})
	}

	resp = AdminCommentListForms{
		Total: total,
		List:  items,
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
			Category:    "normal",
			Type:        "review",
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

// UserList 精确筛选用户列表（学号/姓名/昵称）
func (a *AdminSvc) UserList(info AdminUserListParams) (resp AdminUserForms, err error) {
	var users []model.User
	var total int64

	query := model.DB.Model(&model.User{})

	if info.NetID != "" {
		query = query.Where("netid = ?", info.NetID)
	}
	if info.Name != "" {
		query = query.Where("name = ?", info.Name)
	}
	if info.Nickname != "" {
		query = query.Where("nickname = ?", info.Nickname)
	}
	if info.Status != "" {
		query = query.Where("status = ?", info.Status)
	}
	if info.Level != 0 {
		query = query.Where("level = ?", info.Level)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at ASC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&users).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]UserForm, 0, len(users))
	for _, u := range users {
		resp.List = append(resp.List, UserForm{
			ID:        u.BaseModel.ID,
			NetID:     u.NetID,
			Username:  u.Name,
			Nickname:  u.Nickname,
			AvatarURL: u.AvatarURL,
			Level:     u.Level,
			Status:    u.Status,
		})
	}
	return resp, nil
}

// UpdateAdminLevel 高级管理员调整其他管理员等级（不超过自身等级）
func (a *AdminSvc) UpdateAdminLevel(info AdminUpdateLevelParams) (resp ResponseIS, err error) {
	// ----------------仅 Level >= 3 可操作调整管理员等级----------------
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

	var targetUser model.User
	if err := tx.First(&targetUser, info.ID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("用户不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 不能操作同等级或更高级的管理员
	if targetUser.Level >= info.OperatorLevel {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("无法调整同等级或更高级管理员"), common.LevelErr)
	}

	// oldLevel := targetUser.Level
	// if oldLevel == info.TargetLevel {
	// 	tx.Rollback()
	// 	return resp, common.ErrNew(errors.New("目标等级与当前等级相同"), common.OpErr)
	// }
	// 更新管理员等级(可升可降，但不能超过操作者等级)
	if err := tx.Model(&targetUser).Update("level", info.TargetLevel).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     targetUser.ID,
		Status: "success",
	}
	return resp, nil
}

// SearchUsers 按关键词搜索用户（学号/姓名/昵称）
func (a *AdminSvc) SearchUsers(info AdminSearchUsersParams) (resp AdminUserForms, err error) {
	var users []model.User
	var total int64

	query := model.DB.Model(&model.User{})
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		query = query.Where("netid LIKE ? OR name LIKE ? OR nickname LIKE ?", kw, kw, kw)
	}
	if info.Status != "" {
		query = query.Where("status = ?", info.Status)
	}
	if info.Level != 0 {
		query = query.Where("level = ?", info.Level)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("id ASC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&users).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]UserForm, 0, len(users))
	for _, u := range users {
		resp.List = append(resp.List, UserForm{
			ID:        u.BaseModel.ID,
			NetID:     u.NetID,
			Username:  u.Name,
			Nickname:  u.Nickname,
			AvatarURL: u.AvatarURL,
			Level:     u.Level,
			Status:    u.Status,
		})
	}
	return resp, nil
}

// SetUserStatus 封禁/解封用户，Level 2 可操作 Level 1，Level 3 可操作 Level 1/2，不可操作 Level 3
func (a *AdminSvc) SetUserStatus(info AdminSetUserStatusParams) (resp ResponseIS, err error) {
	if info.OperatorLevel < 2 {
		return resp, common.ErrNew(errors.New("仅管理员可封禁/解封用户"), common.LevelErr)
	}

	var targetUser model.User
	if err := model.DB.First(&targetUser, info.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("用户不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if targetUser.ID == info.OperatorID {
		return resp, common.ErrNew(errors.New("不能封禁/解封自己"), common.ParamErr)
	}

	if targetUser.Level >= info.OperatorLevel {
		return resp, common.ErrNew(errors.New("无法操作同等级或更高级用户"), common.LevelErr)
	}
	if targetUser.Level >= 3 {
		return resp, common.ErrNew(errors.New("无法操作高级管理员"), common.LevelErr)
	}

	// 幂等：相同状态直接返回
	if targetUser.Status == info.Status {
		return ResponseIS{ID: targetUser.ID, Status: targetUser.Status}, nil
	}

	if err := model.DB.Model(&targetUser).Update("status", info.Status).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	return ResponseIS{ID: targetUser.ID, Status: info.Status}, nil
}
