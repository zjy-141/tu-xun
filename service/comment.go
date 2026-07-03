package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Comment struct{}

// Create 创建评论
func (c *Comment) Create(params CreateCommentParams) (resp CreateCommentResponse, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 检查图片是否存在且已审核通过
	var photo model.Photo
	if err := tx.First(&photo, params.PhotoID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}
	if photo.Status != "approved" {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该图片尚未通过审核，暂不可评论"), common.OpErr)
	}

	comment := &model.Comment{
		PhotoID:     params.PhotoID,
		NetID:       params.NetID,
		CommentText: params.CommentText,
		Status:      "pending",
	}

	if err := tx.Create(comment).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return CreateCommentResponse{
		ID:      comment.ID,
		Message: "评论已提交，等待审核",
	}, nil
}

// ListByUser 获取某用户的所有评论
func (c *Comment) ListByUser(params ListUserCommentsParams) (resp ListCommentsResponse, err error) {
	var comments []model.Comment
	var total int64

	query := model.DB.Model(&model.Comment{}).Where("user_id = ?", params.NetID)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	switch params.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "likes_count":
		query = query.Order("likes_count DESC")
	default:
		query = query.Order("created_at DESC")
	}

	if err := query.Preload("User").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&comments).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.Total = total
	resp.Comments = make([]CommentForm, 0, len(comments))
	for _, cm := range comments {
		resp.Comments = append(resp.Comments, CommentForm{
			ID:         cm.ID,
			Content:    cm.CommentText,
			CreatedAt:  cm.CreatedAt,
			LikesCount: cm.LikesCount,
			User: UserBrief{
				ID:        cm.User.ID,
				Name:      cm.User.Name,
				AvatarURL: cm.User.AvatarURL,
			},
		})
	}

	return resp, nil
}

// ListByPhoto 获取某图片下的已审核评论
func (c *Comment) ListByPhoto(params ListPhotoCommentsParams) (resp ListCommentsResponse, err error) {
	var comments []model.Comment
	var total int64

	query := model.DB.Model(&model.Comment{}).
		Where("photo_id = ? AND status = ?", params.PhotoID, "approved")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	switch params.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "likes_count":
		query = query.Order("likes_count DESC")
	default:
		query = query.Order("created_at DESC")
	}
	if err := query.Preload("User").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&comments).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Comments = make([]CommentForm, 0, len(comments))
	for _, cm := range comments {
		resp.Comments = append(resp.Comments, CommentForm{
			ID:         cm.ID,
			Content:    cm.CommentText,
			CreatedAt:  cm.CreatedAt,
			LikesCount: cm.LikesCount,
			User: UserBrief{
				ID:        cm.User.ID,
				Name:      cm.User.Name,
				AvatarURL: cm.User.AvatarURL,
			},
		})
	}

	return resp, nil
}
