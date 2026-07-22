package service

import (
	"errors"
	"fmt"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type MessageSvc struct{}

// ==================== 统一通知 ====================

// List 获取当前用户的通知列表
func (m *MessageSvc) List(info NotificationListParams) (resp NotificationForms, err error) {
	var messages []model.Message
	var total int64

	query := model.DB.Model(&model.Message{}).Where("user_id = ?", info.UserID)

	if info.Category != "" {
		query = query.Where("category = ?", info.Category)
	}
	if info.Type != "" {
		query = query.Where("type = ?", info.Type)
	}
	if info.RelatedType != "" && info.RelatedID > 0 {
		query = query.Where("related_type = ? AND related_id = ?", info.RelatedType, info.RelatedID)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at DESC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&messages).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]NotificationForm, 0, len(messages))
	for _, msg := range messages {
		nf := NotificationForm{
			ID:        msg.ID,
			Category:  msg.Category,
			Type:      msg.Type,
			Title:     msg.Title,
			Content:   msg.Content,
			IsRead:    msg.IsRead,
			CreatedAt: &msg.CreatedAt,
		}
		if msg.SenderID != 0 {
			nf.SenderID = msg.SenderID
		}
		if msg.RelatedID != 0 {
			nf.RelatedID = msg.RelatedID
		}
		if msg.RelatedType != "" {
			nf.RelatedType = msg.RelatedType
		}
		if msg.ExpiresAt != nil {
			nf.ExpiresAt = msg.ExpiresAt
		}
		resp.List = append(resp.List, nf)
	}
	return resp, nil
}

// Detail 获取通知详情
func (m *MessageSvc) Detail(info NotificationGetByIDParams) (resp NotificationDetail, err error) {
	var message model.Message
	if err := model.DB.First(&message, info.NotificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("通知不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = NotificationDetail{
		ID:        message.ID,
		Category:  message.Category,
		Type:      message.Type,
		Title:     message.Title,
		Content:   message.Content,
		IsRead:    message.IsRead,
		CreatedAt: &message.CreatedAt,
	}
	if message.SenderID != 0 {
		resp.SenderID = message.SenderID
	}
	if message.RelatedID != 0 {
		resp.RelatedID = message.RelatedID
	}
	if message.RelatedType != "" {
		resp.RelatedType = message.RelatedType
	}
	if message.ExpiresAt != nil {
		resp.ExpiresAt = message.ExpiresAt
	}

	return resp, nil
}

// MarkAsRead 标记通知为已读
func (m *MessageSvc) MarkAsRead(info NotificationReadParams) (err error) {
	result := model.DB.Model(&model.Message{}).
		Where("id = ? AND user_id = ?", info.NotificationID, info.UserID).
		Update("is_read", true)
	if result.Error != nil {
		return common.ErrNew(result.Error, common.SysErr)
	}
	if result.RowsAffected == 0 {
		return common.ErrNew(errors.New("通知不存在或无权操作"), common.OpErr)
	}
	return nil
}

// GetUnreadCount 获取未读通知数
func (m *MessageSvc) GetUnreadCount(userID int64) (resp NotificationUnreadCount, err error) {
	var count int64
	if err := model.DB.Model(&model.Message{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.Count = count
	return resp, nil
}

// GetGlobalAnnouncement 获取当前用户最新一条未读且未过期的全局公告
func (m *MessageSvc) GetGlobalAnnouncement(userID int64) (resp *NotificationForm, err error) {
	now := time.Now()
	var msg model.Message
	if err := model.DB.Model(&model.Message{}).
		Where("user_id = ? AND type = ? AND is_read = ? AND (expires_at IS NULL OR expires_at > ?)", userID, "global_announcement", false, now).
		Order("created_at DESC, id DESC").
		First(&msg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 无公告，返回 nil 表示 204
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	resp = &NotificationForm{
		ID:        msg.ID,
		Category:  msg.Category,
		Type:      msg.Type,
		Title:     msg.Title,
		Content:   msg.Content,
		IsRead:    msg.IsRead,
		CreatedAt: &msg.CreatedAt,
	}
	if msg.SenderID != 0 {
		resp.SenderID = msg.SenderID
	}
	if msg.RelatedID != 0 {
		resp.RelatedID = msg.RelatedID
	}
	if msg.RelatedType != "" {
		resp.RelatedType = msg.RelatedType
	}
	if msg.ExpiresAt != nil {
		resp.ExpiresAt = msg.ExpiresAt
	}
	return resp, nil
}

// CreateNotification 管理员创建通知
func (m *MessageSvc) CreateNotification(info CreateNotificationRequest) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	switch info.Type {
	case "general":
		// related_type 和 related_id 必须同时提供或同时省略
		hasRelated := info.RelatedType != "" && info.RelatedID > 0
		noRelated := info.RelatedType == "" && info.RelatedID == 0
		if !hasRelated && !noRelated {
			return resp, common.ErrNew(errors.New("related_type 和 related_id 必须同时提供或同时省略"), common.ParamErr)
		}
		if info.ExpiresAt != nil {
			return resp, common.ErrNew(errors.New("general 类型通知不能设置 expires_at"), common.ParamErr)
		}
	case "global_announcement":
		if info.ExpiresAt == nil {
			return resp, common.ErrNew(errors.New("global_announcement 类型通知必须设置 expires_at"), common.ParamErr)
		}
		if !info.ExpiresAt.After(time.Now()) {
			return resp, common.ErrNew(errors.New("expires_at 必须晚于当前时间"), common.ParamErr)
		}
		if info.RelatedType != "" || info.RelatedID != 0 {
			return resp, common.ErrNew(errors.New("global_announcement 类型通知不能带关联对象"), common.ParamErr)
		}
	}

	// 全局公告发给所有用户
	if info.Type == "global_announcement" {
		var userIDs []int64
		if err := tx.Model(&model.User{}).Pluck("id", &userIDs).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
		for _, uid := range userIDs {
			msg := &model.Message{
				UserID:      uid,
				SenderID:    1,
				Category:    "normal",
				Type:        info.Type,
				Title:       info.Title,
				Content:     info.Content,
				RelatedID:   info.RelatedID,
				RelatedType: info.RelatedType,
				IsRead:      false,
				ExpiresAt:   info.ExpiresAt,
			}
			if err := tx.Create(msg).Error; err != nil {
				return resp, common.ErrNew(err, common.SysErr)
			}
		}
	} else {
		// general 类型：发给所有用户的通知
		var userIDs []int64
		if err := tx.Model(&model.User{}).Pluck("id", &userIDs).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
		for _, uid := range userIDs {
			msg := &model.Message{
				UserID:      uid,
				SenderID:    1,
				Category:    "normal",
				Type:        info.Type,
				Title:       info.Title,
				Content:     info.Content,
				RelatedID:   info.RelatedID,
				RelatedType: info.RelatedType,
				IsRead:      false,
			}
			if err := tx.Create(msg).Error; err != nil {
				return resp, common.ErrNew(err, common.SysErr)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}

	return ResponseIS{ID: 0, Status: "published"}, nil
}

// SendLikeNotification 发送点赞通知（互动消息，只投递给目标用户）
func (m *MessageSvc) SendLikeNotification(likerID int64, targetType string, targetID int64, ownerID int64) {
	if likerID == ownerID || ownerID == 0 {
		return // 不给自己发通知
	}
	msg := &model.Message{
		UserID:      ownerID,
		SenderID:    likerID,
		Category:    "interaction",
		Type:        "like",
		Title:       "有人点赞了你",
		Content:     fmt.Sprintf("有人点赞了你的%s，快去看看吧！", relatedTypeNames[targetType]),
		RelatedID:   targetID,
		RelatedType: targetType,
		IsRead:      false,
	}
	_ = model.DB.Create(msg).Error
}

var relatedTypeNames = map[string]string{
	"photo":   "图片投稿",
	"attempt": "答题",
	"comment": "评论",
}
