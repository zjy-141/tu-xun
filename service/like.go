package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type LikeSvc struct{}

// SetLike 切换点赞状态（PUT 语义，无请求体），返回操作后的状态和计数
func (l *LikeSvc) SetLike(userID int64, targetType string, targetID int64) (resp LikeResult, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 检查目标是否存在
	if err := l.checkTarget(tx, targetType, targetID); err != nil {
		tx.Rollback()
		return resp, err
	}

	// 查是否已点赞
	var existing model.Like
	result := tx.Where("user_id = ? AND target_type = ? AND target_id = ?",
		userID, targetType, targetID).First(&existing)

	if result.Error == nil {
		// 已点赞 → 取消
		if err := tx.Unscoped().Delete(&existing).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
		l.decrCounter(tx, targetType, targetID)
		resp.Liked = false
	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 未点赞 → 点赞
		like := &model.Like{
			UserID:     userID,
			TargetType: targetType,
			TargetID:   targetID,
		}
		if err := tx.Create(like).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
		l.incrCounter(tx, targetType, targetID)
		resp.Liked = true
	} else {
		tx.Rollback()
		return resp, common.ErrNew(result.Error, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	// 点赞成功时发送通知给目标所有者
	if resp.Liked {
		ownerID := l.getOwnerID(targetType, targetID)
		if ownerID > 0 && ownerID != userID {
			msgSvc := MessageSvc{}
			photoID := l.getPhotoIDForTarget(targetType, targetID)
			msgSvc.SendInteraction(userID, targetType, targetID, ownerID, photoID)
		}
	}

	resp.LikesCount, err = l.getCount(targetType, targetID)
	if err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	return resp, nil
}

// checkTarget 检查点赞目标是否存在
func (l *LikeSvc) checkTarget(tx *gorm.DB, targetType string, targetID int64) error {
	switch targetType {
	case "photo":
		var p model.Photo
		if err := tx.First(&p, targetID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return common.ErrNew(errors.New("图片不存在"), common.OpErr)
			}
			return common.ErrNew(err, common.SysErr)
		}
	case "comment":
		var c model.Comment
		if err := tx.First(&c, targetID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return common.ErrNew(errors.New("评论不存在"), common.OpErr)
			}
			return common.ErrNew(err, common.SysErr)
		}
	case "attempt":
		var a model.Attempt
		if err := tx.First(&a, targetID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return common.ErrNew(errors.New("答题记录不存在"), common.OpErr)
			}
			return common.ErrNew(err, common.SysErr)
		}
	default:
		return common.ErrNew(errors.New("无效的点赞目标类型"), common.ParamErr)
	}
	return nil
}

// incrCounter 点赞计数+1
func (l *LikeSvc) incrCounter(tx *gorm.DB, targetType string, targetID int64) {
	tableName := targetType
	if targetType == "attempt" {
		tableName = "attempt"
	}
	tx.Table(tableName).Where("id = ?", targetID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + 1"))
}

// decrCounter 点赞计数-1
func (l *LikeSvc) decrCounter(tx *gorm.DB, targetType string, targetID int64) {
	tableName := targetType
	if targetType == "attempt" {
		tableName = "attempt"
	}
	tx.Table(tableName).Where("id = ?", targetID).
		UpdateColumn("likes_count", gorm.Expr("likes_count - 1"))
}

// getCount 获取当前点赞数
func (l *LikeSvc) getCount(targetType string, targetID int64) (resp int64, err error) {
	var count int64
	if err := model.DB.Model(&model.Like{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Count(&count).Error; err != nil {
		return resp, err
	}
	return count, nil
}

// getOwnerID 获取目标内容的所有者ID
func (l *LikeSvc) getOwnerID(targetType string, targetID int64) int64 {
	switch targetType {
	case "photo":
		var p model.Photo
		if err := model.DB.First(&p, targetID).Error; err == nil {
			return p.UserID
		}
	case "comment":
		var c model.Comment
		if err := model.DB.First(&c, targetID).Error; err == nil {
			return c.UserID
		}
	case "attempt":
		var a model.Attempt
		if err := model.DB.First(&a, targetID).Error; err == nil {
			return a.UserID
		}
	}
	return 0
}

// getPhotoIDForTarget 获取目标内容关联的题目ID
func (l *LikeSvc) getPhotoIDForTarget(targetType string, targetID int64) int64 {
	switch targetType {
	case "photo":
		return targetID
	case "comment":
		var c model.Comment
		if err := model.DB.Select("photo_id").First(&c, targetID).Error; err == nil {
			return c.PhotoID
		}
	case "attempt":
		var a model.Attempt
		if err := model.DB.Select("photo_id").First(&a, targetID).Error; err == nil {
			return a.PhotoID
		}
	}
	return 0
}
