package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Attempt struct{}

// Create 提交答题
func (a *Attempt) Create(info CreateAttemptParams) (*SubmitAttemptResponse, error) {
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

	// // 检查是否已有待审核的答题记录（同一用户同一图片）
	// var existAttempt model.Attempt
	// if err := tx.Where("photo_id = ? AND user_id = ? AND status = ?", info.PhotoID, info.UserID, "pending").
	// 	First(&existAttempt).Error; err == nil {
	// 	tx.Rollback()
	// 	return nil, common.ErrNew(errors.New("您已提交过答题，请等待管理员审核"), common.OpErr)
	// }

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
		IsWinner:        false,
		Status:          "pending",
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

	return &SubmitAttemptResponse{
		AttemptID: attempt.ID,
		PhotoID:   attempt.PhotoID,
		Status:    attempt.Status,
		Message:   "已提交，等待管理员审核。若审核通过且本题尚未被破解，您将获得奖品。",
	}, nil
}

// ListByUser 获取某用户的所有答题记录（个人主页用）
func (a *Attempt) AttemptShow(info AttemptShowParams) (resp ListAttemptsResponse, err error) {

	var total int64
	var attempts []model.Attempt
	query := model.DB.Model(&model.Attempt{}).Where("user_id = ?", info.UserID)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	switch info.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "attempts_count":
		query = query.Order("attempts_count DESC")
	case "likes_count":
		query = query.Order("likes_count DESC")
	default:
		query = query.Order("created_at DESC")
	}
	if err := query.Scopes(model.Paginate(info.PagerForm)).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.Total = total
	resp.Attempts = make([]AttemptForm, 0, len(attempts))
	for _, at := range attempts {
		resp.Attempts = append(resp.Attempts, AttemptForm{
			ID:              at.ID,
			ImageURL:        at.ImageURL,
			GuessedLocation: at.GuessedLocation,
			LikesCount:      at.LikesCount,
			CreatedAt:       at.CreatedAt,
			User: UserBrief{
				ID:        at.User.ID,
				Name:      at.User.Name,
				AvatarURL: at.User.AvatarURL,
			},
		})
	}
	return resp, nil
}

// ListByPhoto 获取某图片下的已审核答题记录
func (a *Attempt) ListByPhoto(params ListPhotoAttemptsParams) (resp ListAttemptsResponse, err error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{}).
		Where("photo_id = ? AND status = ?", params.PhotoID, "approved")

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
			ID:              at.ID,
			ImageURL:        at.ImageURL,
			GuessedLocation: at.GuessedLocation,
			LikesCount:      at.LikesCount,
			CreatedAt:       at.CreatedAt,
			User: UserBrief{
				ID:        at.User.ID,
				Name:      at.User.Name,
				AvatarURL: at.User.AvatarURL,
			},
		})
	}

	return resp, nil
}
