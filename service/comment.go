package service

import (
	"errors"
	"time"

	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/model"
	"tu-xun/pkg/sensitive"

	"gorm.io/gorm"
)

type CommentSvc struct{}

// Create 创建评论
func (c *CommentSvc) Create(params CommentCreateParams) (ResponseIS, error) {
	tx := model.DB.Begin()
	var err error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	// 检查图片是否存在且已审核通过
	var photo model.Photo
	if err = tx.Preload("Activity").First(&photo, params.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ResponseIS{}, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return ResponseIS{}, common.ErrNew(err, common.SysErr)
	}
	if photo.Status != "approved" {
		return ResponseIS{}, common.ErrNew(errors.New("该图片尚未通过审核，暂不可评论"), common.OpErr)
	}

	// 活动必须是 active 或 ended（排除 not_started）
	if photo.Activity.StartTime != nil && photo.Activity.StartTime.After(time.Now()) {
		return ResponseIS{}, common.ErrNew(errors.New("该活动尚未开始，暂不可评论"), common.OpErr)
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

	if err = tx.Create(comment).Error; err != nil {
		return ResponseIS{}, common.ErrNew(err, common.SysErr)
	}

	// 自动审核：在事务内完成状态更新
	if config.Config.AUTO_APPROVAL == "attemptAndComment" || config.Config.AUTO_APPROVAL == "all" {
		now := time.Now()
		comment.Status = status
		comment.ReviewedAt = &now

		if status == "rejected" {
			comment.RejectReason = "自动审核中"
		}

		// 持久化审核状态和审核时间
		if err := tx.Save(&comment).Error; err != nil {
			tx.Rollback()
			return ResponseIS{}, common.ErrNew(err, common.SysErr)
		}
	}

	if err = tx.Commit().Error; err != nil {
		return ResponseIS{}, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return ResponseIS{
		ID:     comment.ID,
		Status: comment.Status,
	}, nil
}

// ListByPhoto 获取某图片下的已审核评论
func (c *CommentSvc) ListByPhoto(params CommentListParams) (CommentItemPage, error) {
	var comments []model.Comment
	var total int64

	query := model.DB.Model(&model.Comment{}).
		Where("photo_id = ? AND status = ?", params.PhotoID, "approved")

	if err := query.Count(&total).Error; err != nil {
		return CommentItemPage{}, common.ErrNew(err, common.SysErr)
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
		return CommentItemPage{}, common.ErrNew(err, common.SysErr)
	}

	resp := CommentItemPage{
		Total: total,
		List:  make([]CommentItem, 0, len(comments)),
	}
	for _, cm := range comments {
		resp.List = append(resp.List, CommentItem{
			ID: cm.ID,
			Author: UserBrief{
				ID:        cm.User.ID,
				Nickname:  cm.User.Nickname,
				Avatar: cm.User.AvatarURL,
			},
			Content:    cm.CommentText,
			LikesCount: cm.LikesCount,
			CreatedAt:  &cm.CreatedAt,
		})
	}
	return resp, nil
}

// Delete 删除评论
func (c *CommentSvc) Delete(params CommentDeleteParams) (ResponseIS, error) {
	tx := model.DB.Begin()
	var err error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	// 检查评论是否存在
	var comment model.Comment
	if err = tx.First(&comment, params.CommentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ResponseIS{}, common.ErrNew(errors.New("评论不存在"), common.OpErr)
		}
		return ResponseIS{}, common.ErrNew(err, common.SysErr)
	}
	if comment.UserID != params.UserID && params.Level < 2 {
		return ResponseIS{}, common.ErrNew(errors.New("无权限删除该评论"), common.AuthErr)
	}

	if err = tx.Delete(&comment).Error; err != nil {
		return ResponseIS{}, common.ErrNew(err, common.SysErr)
	}

	if err = tx.Commit().Error; err != nil {
		return ResponseIS{}, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return ResponseIS{
		ID:     params.CommentID,
		Status: "deleted",
	}, nil
}
