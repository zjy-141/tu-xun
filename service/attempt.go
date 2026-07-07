package service

import (
	"errors"
	"time"
	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/model"

	"gorm.io/gorm"
)

type AttemptSvc struct{}

// Create 提交答题
func (a *AttemptSvc) Create(info AttemptCreateParams) (resp ResponseIS, err error) {
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
			tx.Rollback()
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}
	if photo.Status != "approved" {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该图片尚未通过审核，暂不可答题"), common.OpErr)
	}

	// 检查是否已有待审核的答题记录（同一用户同一图片）
	var existtotal int64
	if err := tx.Where("photo_id = ? AND user_id = ?", info.PhotoID, info.UserID).
		Count(&existtotal).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}
	if existtotal > 100 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("您有过多的答题记录待审核，请耐心等待管理员审核"), common.ParamErr)
	}

	// 保存答题图片（仅缩略图）
	imageURL, err := saveThumbnailOnly(info.ImageFile, "attempts")
	if err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	status := "pending"
	if config.Config.AUTO_APPROVAL == "all" {
		//自动审核
	}

	attempt := &model.Attempt{
		PhotoID:     info.PhotoID,
		UserID:      info.UserID,
		CommentText: info.CommentText,
		ImageURL:    imageURL,
		Longitude:   info.Longitude,
		Latitude:    info.Latitude,
		LikesCount:  0,
		Status:      status,
	}

	if err := tx.Create(attempt).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 更新图片的答题次数
	tx.Model(&photo).UpdateColumn("attempts_count", gorm.Expr("attempts_count + 1"))

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     attempt.ID,
		Status: attempt.Status,
	}

	return resp, nil
}

// // ListByUser 获取某用户的所有答题记录（个人主页用）
// func (a *AttemptSvc) AttemptShow(info AttemptShowParams) (resp ListAttemptsResponse, err error) {

// 	var total int64
// 	var attempts []model.Attempt
// 	query := model.DB.Model(&model.Attempt{}).Where("user_id = ?", info.UserID)

// 	if err := query.Count(&total).Error; err != nil {
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	switch info.SortBy {
// 	case "created_at":
// 		query = query.Order("created_at DESC")
// 	case "attempts_count":
// 		query = query.Order("attempts_count DESC")
// 	case "likes_count":
// 		query = query.Order("likes_count DESC")
// 	default:
// 		query = query.Order("created_at DESC")
// 	}
// 	if err := query.Scopes(model.Paginate(info.PagerForm)).
// 		Find(&attempts).Error; err != nil {
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}
// 	resp.Total = total
// 	resp.Attempts = make([]AttemptForm, 0, len(attempts))
// 	for _, at := range attempts {
// 		resp.Attempts = append(resp.Attempts, AttemptForm{
// 			ID:              at.ID,
// 			ImageURL:        at.ImageURL,
// 			GuessedLocation: at.GuessedLocation,
// 			LikesCount:      at.LikesCount,
// 			CreatedAt:       at.CreatedAt,
// 			User: UserBrief{
// 				ID:        at.User.ID,
// 				Name:      at.User.Name,
// 				AvatarURL: at.User.AvatarURL,
// 			},
// 		})
// 	}
// 	return resp, nil
// }

// ListByPhoto 获取某图片下的已审核答题记录
func (a *AttemptSvc) ListByPhoto(params PhotoAttemptsListParams) (resp AttemptForms, err error) {
	var attempts []model.Attempt
	var total int64
	// 查询已审核通过的答题记录，且排除未破解的记录
	query := model.DB.Model(&model.Attempt{}).
		Where("photo_id = ? AND status = ? AND solved != ?", params.PhotoID, "approved", 0)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	switch params.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "attempts_count":
		query = query.Order("attempts_count DESC")
	case "likes_count":
		query = query.Order("likes_count DESC")
	default:
		query = query.Order("created_at DESC")
	}
	if err := query.Preload("User").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Attempts = make([]AttemptForm, 0, len(attempts))
	for _, at := range attempts {
		resp.Attempts = append(resp.Attempts, AttemptForm{
			ID: at.ID,
			Author: UserBrief{
				ID:        at.User.ID,
				Nickname:  at.User.Nickname,
				AvatarURL: at.User.AvatarURL,
			},
			CommentText: at.CommentText,
			ImageURL:    at.ImageURL,
			LikesCount:  at.LikesCount,
			CreatedAt:   at.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return resp, nil
}

// GetUserRank 获取用户在指定图片答题中的排名（按最早答对时间）
func (a *ActivitySvc) GetUserRank(userID int64, photoID int64) (rank int, err error) {
	// 1. 获取该用户最早答对时间（假设不限定活动，如果有活动可加条件）
	var firstTime time.Time
	if err := model.DB.Model(&model.Attempt{}).
		Select("MIN(created_at)").
		Where("user_id = ? AND photo_id = ? ANDsolved = 1", userID, photoID).
		Scan(&firstTime).Error; err != nil {
		return 0, common.ErrNew(err, common.SysErr)
	}
	if firstTime.IsZero() {
		return 0, nil // 未答对，无排名
	}

	// 2. 统计有多少不同用户的【最早答对时间】早于该用户
	if err := model.DB.Raw(`
        SELECT COUNT(DISTINCT user_id) + 1
        FROM attempts
        WHERE solved = 1 AND photo_id = ?
          AND created_at < ?
    `, photoID, firstTime).Scan(&rank).Error; err != nil {
		return 0, common.ErrNew(err, common.SysErr)
	}
	return rank, nil
}
