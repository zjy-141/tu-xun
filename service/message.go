package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type MessageSvc struct{}

// ==================== helpers ====================

var whitespaceRe = regexp.MustCompile(`\s+`)

// generateContentPreview 从 Content 生成摘要：
//  1. 将字面 [image] 替换为空格
//  2. 将换行、制表符、连续 Unicode 空白字符折叠为单个半角空格，并去掉首尾空白
//  3. 取前 50 个 Unicode 码点（符文），不加省略号
func generateContentPreview(content string) string {
	s := strings.ReplaceAll(content, "[image]", " ")
	s = whitespaceRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > 50 {
		s = string(runes[:50])
	}
	return s
}

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
	if info.Keyword != "" {
		keyword := "%" + info.Keyword + "%"
		if id, parseErr := strconv.ParseInt(info.Keyword, 10, 64); parseErr == nil {
			query = query.Where("id = ? OR title LIKE ? OR content LIKE ?", id, keyword, keyword)
		} else {
			query = query.Where("title LIKE ? OR content LIKE ?", keyword, keyword)
		}
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
	resp.List = make([]NotificationListItem, 0, len(messages))
	for _, msg := range messages {
		resp.List = append(resp.List, NotificationListItem{
			ID:             msg.ID,
			Category:       msg.Category,
			Type:           msg.Type,
			Title:          msg.Title,
			ContentPreview: generateContentPreview(msg.Content),
			IsRead:         msg.IsRead,
			CreatedAt:      &msg.CreatedAt,
		})
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
		ImageURL:  message.ImageURL,
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
func (m *MessageSvc) GetGlobalAnnouncement(userID int64) (resp *NotificationDetail, err error) {
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

	resp = &NotificationDetail{
		ID:        msg.ID,
		SenderID:  msg.SenderID,
		Category:  msg.Category,
		Type:      msg.Type,
		Title:     msg.Title,
		Content:   msg.Content,
		ImageURL:  msg.ImageURL,
		IsRead:    msg.IsRead,
		CreatedAt: &msg.CreatedAt,
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

// CreateNotification 管理员创建通知（multipart/form-data）
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

	// 上传图片（可选）
	var imageURL string
	if info.ImageFile != nil {
		imageURL, err = OSSClient.UploadFile(info.ImageFile, "notifications")
		if err != nil {
			return resp, err
		}
	}

	// 发给所有用户
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
			ImageURL:    imageURL,
			RelatedID:   info.RelatedID,
			RelatedType: info.RelatedType,
			IsRead:      false,
			ExpiresAt:   info.ExpiresAt,
		}
		if err := tx.Create(msg).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}

	return ResponseIS{ID: 0, Status: "published"}, nil
}

// UpdateNotification 管理员更新通知
func (m *MessageSvc) UpdateNotification(info UpdateNotificationRequest) (resp ResponseIS, err error) {
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

	var msg model.Message
	if err := tx.First(&msg, info.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("通知不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if msg.Category != "normal" {
		return resp, common.ErrNew(errors.New("互动消息不可更新"), common.OpErr)
	}

	updates := map[string]interface{}{}

	// ---- type ----
	if info.Type != "" && info.Type != msg.Type {
		if info.Type == "general" {
			// global_announcement → general：清除过期时间
			updates["expires_at"] = nil
		} else if info.Type == "global_announcement" {
			// general → global_announcement：必须提供过期时间
			if info.ExpiresAt == nil {
				return resp, common.ErrNew(errors.New("切换为 global_announcement 类型必须设置 expires_at"), common.ParamErr)
			}
		}
		updates["type"] = info.Type
	}

	// ---- title ----
	if info.Title != "" {
		updates["title"] = info.Title
	}

	// ---- content ----
	if info.Content != "" {
		updates["content"] = info.Content
	}

	// ---- image ----
	if info.ImageFile != nil && info.RemoveImage {
		return resp, common.ErrNew(errors.New("不能同时提供 image_file 和 remove_image"), common.ParamErr)
	}
	if info.ImageFile != nil {
		imageURL, uploadErr := OSSClient.UploadFile(info.ImageFile, "notifications")
		if uploadErr != nil {
			return resp, uploadErr
		}
		updates["image_url"] = imageURL
	}
	if info.RemoveImage {
		updates["image_url"] = ""
	}

	// ---- relation ----
	hasRelation := info.RelatedType != "" || info.RelatedID > 0
	if hasRelation && info.RemoveRelation {
		return resp, common.ErrNew(errors.New("不能同时提供 related_type/related_id 和 remove_relation"), common.ParamErr)
	}
	if info.RemoveRelation {
		updates["related_type"] = ""
		updates["related_id"] = 0
	} else if hasRelation {
		if info.RelatedType == "" || info.RelatedID == 0 {
			return resp, common.ErrNew(errors.New("related_type 和 related_id 必须同时提供"), common.ParamErr)
		}
		updates["related_type"] = info.RelatedType
		updates["related_id"] = info.RelatedID
	}

	// ---- expires_at ----
	if info.ExpiresAt != nil {
		newType := msg.Type
		if t, ok := updates["type"]; ok {
			newType = t.(string)
		}
		if newType != "global_announcement" {
			return resp, common.ErrNew(errors.New("仅 global_announcement 类型可设置 expires_at"), common.ParamErr)
		}
		if !info.ExpiresAt.After(time.Now()) {
			return resp, common.ErrNew(errors.New("expires_at 必须晚于当前时间"), common.ParamErr)
		}
		updates["expires_at"] = info.ExpiresAt
	}

	if len(updates) == 0 {
		return ResponseIS{ID: info.ID, Status: "unchanged"}, nil
	}

	if err := tx.Model(&msg).Updates(updates).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
	}

	return ResponseIS{ID: info.ID, Status: "updated"}, nil
}

// DeleteNotification 管理员删除通知（软删除）
func (m *MessageSvc) DeleteNotification(notificationID int64) (resp ResponseIS, err error) {
	var msg model.Message
	if err := model.DB.First(&msg, notificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("通知不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if msg.Category != "normal" {
		return resp, common.ErrNew(errors.New("互动消息不可删除"), common.OpErr)
	}

	if err := model.DB.Delete(&msg).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	return ResponseIS{ID: notificationID, Status: "deleted"}, nil
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
