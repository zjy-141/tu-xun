package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/model"
	"tu-xun/pkg/sensitive"

	"gorm.io/gorm"
)

type CommentSvc struct{}

// Create 创建评论
func (c *CommentSvc) Create(params CommentCreateParams) (resp ResponseIS, err error) {
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

	status := "pending"
	if config.Config.AUTO_APPROVAL == "comment" || config.Config.AUTO_APPROVAL == "attemptAndComment" || config.Config.AUTO_APPROVAL == "all" {
		// 敏感词检测：包含敏感词则拒绝，否则自动通过
		if sensitive.Detect(params.CommentText) {
			status = "rejected"
		} else {
			status = "approved"
		}
	}

	comment := &model.Comment{
		PhotoID:     params.PhotoID,
		UserID:      params.UserID,
		CommentText: params.CommentText,
		Status:      status,
	}

	if err := tx.Create(comment).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	if config.Config.AUTO_APPROVAL == "attemptAndComment" || config.Config.AUTO_APPROVAL == "all" {
		//自动发放积分并通知
		a := AdminSvc{}
		info := AdminReviewCommentParams{
			CommentID:    comment.ID,
			Action:       status,
			RejectReason: "自动审核中",
		}
		resp, err := a.ReviewComment(info)
		if err != nil {
			return resp, err
		}

	}
	resp = ResponseIS{
		ID:     comment.ID,
		Status: comment.Status,
	}
	return resp, nil
}

// // ListByUser 获取某用户的所有评论
// func (c *CommentSvc) ListByUser(params ListUserCommentsParams) (resp CommentForms, err error) {
// 	var comments []model.Comment
// 	var total int64

// 	query := model.DB.Model(&model.Comment{}).Where("user_id = ?", params.UserID)

// 	if err := query.Count(&total).Error; err != nil {
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	switch params.SortBy {
// 	case "created_at":
// 		query = query.Order("created_at DESC")
// 	case "likes_count":
// 		query = query.Order("likes_count DESC")
// 	default:
// 		query = query.Order("created_at DESC")
// 	}

// 	if err := query.Preload("User").
// 		Scopes(model.Paginate(params.PagerForm)).
// 		Find(&comments).Error; err != nil {
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}
// 	resp.Total = total
// 	resp.Comments = make([]CommentForm, 0, len(comments))
// 	for _, cm := range comments {
// 		resp.Comments = append(resp.Comments, CommentForm{
// 			ID:         cm.ID,
// 			Content:    cm.CommentText,
// 			CreatedAt:  cm.CreatedAt,
// 			LikesCount: cm.LikesCount,
// 			User: UserBrief{
// 				ID:        cm.User.ID,
// 				Name:      cm.User.Name,
// 				AvatarURL: cm.User.AvatarURL,
// 			},
// 		})
// 	}

// 	return resp, nil
// }

// ListByPhoto 获取某图片下的已审核评论
func (c *CommentSvc) ListByPhoto(params PhotoCommentsListParams) (resp CommentForms, err error) {
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
			ID: cm.ID,
			Author: UserBrief{
				ID:        cm.User.ID,
				Nickname:  cm.User.Nickname,
				AvatarURL: cm.User.AvatarURL,
			},
			CommentText: cm.CommentText,
			LikesCount:  cm.LikesCount,
			CreatedAt:   cm.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}

// Delete 删除评论
func (c *CommentSvc) Delete(params CommentDeleteParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 检查评论是否存在且已审核通过
	var comment model.Comment
	if err := tx.First(&comment, params.CommentID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("评论不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}
	if comment.UserID != params.UserID && params.Level < 2 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("无权限删除该评论"), common.AuthErr)
	}

	if err := tx.Delete(&comment).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     params.CommentID,
		Status: "deleted",
	}
	return resp, nil
}
