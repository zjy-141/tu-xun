package service

import (
	"errors"
	"fmt"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/model"

	"gorm.io/gorm"
)

type MessageSvc struct{}

// ==================== 通知消息 ====================

// ListMyMessages 获取当前用户的通知消息列表
func (m *MessageSvc) ListMyMessages(info ListMessageParams) (resp ListMessagesResponse, err error) {
	var messages []model.Message
	var total int64

	query := model.DB.Model(&model.Message{}).Where("user_id = ? AND type != ?", info.NetID, "chat")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at DESC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&messages).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Messages = make([]MessageItem, 0, len(messages))
	for _, msg := range messages {
		resp.Messages = append(resp.Messages, MessageItem{
			ID:          msg.ID,
			Type:        msg.Type,
			Title:       msg.Title,
			Content:     msg.Content,
			RelatedID:   msg.RelatedID,
			RelatedType: msg.RelatedType,
			IsRead:      msg.IsRead,
			CreatedAt:   msg.CreatedAt,
		})
	}
	return resp, nil
}

// MarkAsRead 标记消息为已读
func (m *MessageSvc) MarkAsRead(messageID int64, NetID int64) error {
	result := model.DB.Model(&model.Message{}).
		Where("id = ? AND user_id = ?", messageID, NetID).
		Update("is_read", true)
	if result.Error != nil {
		return common.ErrNew(result.Error, common.SysErr)
	}
	if result.RowsAffected == 0 {
		return common.ErrNew(errors.New("消息不存在或无权操作"), common.OpErr)
	}
	return nil
}

// GetUnreadCount 获取未读通知数
func (m *MessageSvc) GetUnreadCount(NetID int64) (resp UnreadCountResponse, err error) {
	var count int64
	if err := model.DB.Model(&model.Message{}).
		Where("user_id = ? AND is_read = ? AND type != ?", NetID, false, "chat").
		Count(&count).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.Count = count
	return resp, nil
}

// SendReviewMessage 发送审核结果消息
func (m *MessageSvc) SendReviewMessage(NetID int64, action string, relatedID int64, relatedType string, rejectReason string) error {
	var msgType, title, content string

	switch action {
	case "approve":
		msgType = "review_approved"
		title = fmt.Sprintf("您的%s已通过审核", relatedTypeNames[relatedType])
		content = fmt.Sprintf("恭喜！您提交的%s已通过审核。", relatedTypeNames[relatedType])
	case "reject":
		msgType = "review_rejected"
		title = fmt.Sprintf("您的%s未通过审核", relatedTypeNames[relatedType])
		content = fmt.Sprintf("您提交的%s未通过审核。拒绝原因：%s", relatedTypeNames[relatedType], rejectReason)
	default:
		return nil
	}

	msg := &model.Message{
		NetID:       NetID,
		SenderID:    1, // 系统消息
		Type:        msgType,
		Title:       title,
		Content:     content,
		RelatedID:   relatedID,
		RelatedType: relatedType,
		IsRead:      false,
	}

	if err := model.DB.Create(msg).Error; err != nil {
		return common.ErrNew(err, common.SysErr)
	}
	return nil
}

// SendLikeNotification 发送点赞通知
func (m *MessageSvc) SendLikeNotification(likerID int64, targetType string, targetID int64, ownerID int64) {
	if likerID == ownerID {
		return // 不给自己发通知
	}
	msg := &model.Message{
		NetID:       ownerID,
		SenderID:    1,
		Type:        "like",
		Title:       "有人点赞了你",
		Content:     fmt.Sprintf("有人点赞了你的%s，快去看看吧！", relatedTypeNames[targetType]),
		RelatedID:   targetID,
		RelatedType: targetType,
		IsRead:      false,
	}
	_ = model.DB.Create(msg).Error
}

// SendAttemptNotification 发送答题通知
func (m *MessageSvc) SendAttemptNotification(submitterID int64, photoID int64, ownerID int64) {
	if submitterID == ownerID {
		return // 不给自己发通知
	}
	msg := &model.Message{
		NetID:       ownerID,
		SenderID:    1,
		Type:        "attempt",
		Title:       "有人挑战了你的图片",
		Content:     "有人提交了新的答题，等待管理员审核。",
		RelatedID:   photoID,
		RelatedType: "photo",
		IsRead:      false,
	}
	_ = model.DB.Create(msg).Error
}

var relatedTypeNames = map[string]string{
	"photo":   "图片投稿",
	"attempt": "答题",
	"comment": "评论",
}

// ==================== 会话（微信风格聊天） ====================

// ListConversations 获取会话列表（微信首页：系统通知 + 用户聊天）
func (m *MessageSvc) ListConversations(NetID int64) (resp ListConversationsResponse, err error) {
	var rows []struct {
		PartnerID int64
	}
	if err := model.DB.Raw(`
		SELECT DISTINCT partner_id FROM (
			SELECT 1 AS partner_id FROM message WHERE user_id = ? AND type != 'chat'
			UNION
			SELECT sender_id AS partner_id FROM message WHERE user_id = ? AND type = 'chat'
			UNION
			SELECT user_id AS partner_id FROM message WHERE sender_id = ? AND type = 'chat'
		) t
	`, NetID, NetID, NetID).Scan(&rows).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	conversations := make([]ConversationItem, 0, len(rows))
	for _, row := range rows {
		var item ConversationItem
		var itemErr error
		if row.PartnerID == 1 {
			item, itemErr = m.buildSystemConversation(NetID)
		} else {
			item, itemErr = m.buildConversationItem(NetID, row.PartnerID)
		}
		if itemErr != nil {
			logger.Errorf("构建会话项失败: NetID=%d, partnerID=%d, error=%v\n", NetID, row.PartnerID, itemErr)
			continue
		}
		conversations = append(conversations, item)
	}

	// 按最新消息时间倒序
	sortConversations(conversations)
	resp.Conversations = conversations
	return resp, nil
}

// buildSystemConversation 构建系统通知虚拟会话
func (m *MessageSvc) buildSystemConversation(NetID int64) (item ConversationItem, err error) {
	// 查最后一条消息
	var lastMsg model.Message
	err = model.DB.Where("user_id = ? AND type != ?", NetID, "chat").
		Order("created_at DESC").First(&lastMsg).Error
	if err != nil {
		return item, common.ErrNew(err, common.SysErr)
	}

	// 查未读数（对方发给我的未读消息）
	var unread int64
	model.DB.Model(&model.Message{}).
		Where("user_id = ? AND is_read = ? AND type != ?", NetID, false, "chat").
		Count(&unread)

	item = ConversationItem{
		PartnerID:     1,
		PartnerName:   "系统通知",
		PartnerAvatar: "",
		LastContent:   lastMsg.Title,
		LastTime:      lastMsg.CreatedAt,
		UnreadCount:   unread,
	}
	return item, nil
}

// sortConversations 按 LastTime 降序冒泡排序
func sortConversations(items []ConversationItem) {
	for i := range items {
		for j := i + 1; j < len(items); j++ {
			if items[j].LastTime.After(items[i].LastTime) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// buildConversationItem 构建单个会话项
func (m *MessageSvc) buildConversationItem(NetID, partnerID int64) (item ConversationItem, err error) {
	// 查对方信息
	var partner model.User
	if err := model.DB.First(&partner, partnerID).Error; err != nil {
		return item, common.ErrNew(err, common.OpErr)
	}

	// 查最后一条消息
	var lastMsg model.Message
	model.DB.Where(
		"((user_id = ? AND sender_id = ?) OR (user_id = ? AND sender_id = ?)) AND type = ?",
		NetID, partnerID, partnerID, NetID, "chat",
	).Order("created_at DESC").First(&lastMsg)

	// 查未读数（对方发给我的未读消息）
	var unread int64
	model.DB.Model(&model.Message{}).Where(
		"user_id = ? AND sender_id = ? AND type = ? AND is_read = ?",
		NetID, partnerID, "chat", false,
	).Count(&unread)

	item = ConversationItem{
		PartnerID:     partner.ID,
		PartnerName:   partner.Name,
		PartnerAvatar: partner.AvatarURL,
		LastContent:   lastMsg.Content,
		LastTime:      lastMsg.CreatedAt,
		UnreadCount:   unread,
	}
	return item, nil
}

// GetConversation 获取对话详情（微信聊天窗口 / 系统通知列表）
func (m *MessageSvc) GetConversation(info GetConversationParams) (resp ConversationDetailResponse, err error) {
	// --- partner_id=1 → 系统通知 ---
	if info.PartnerID == 1 {
		return m.getSystemMessages(info)
	}

	// --- partner_id>0 → 用户聊天 ---
	// 查对方信息
	var partner model.User
	if err := model.DB.First(&partner, info.PartnerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("用户不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 标记对方发来的未读消息为已读
	model.DB.Model(&model.Message{}).Where(
		"user_id = ? AND sender_id = ? AND type = ? AND is_read = ?",
		info.NetID, info.PartnerID, "chat", false,
	).Update("is_read", true)

	// 查双方对话
	var messages []model.Message
	var total int64

	query := model.DB.Model(&model.Message{}).Where(
		"((user_id = ? AND sender_id = ?) OR (user_id = ? AND sender_id = ?)) AND type = ?",
		info.NetID, info.PartnerID, info.PartnerID, info.NetID, "chat",
	)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at ASC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&messages).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	chatMsgs := make([]ChatMessage, 0, len(messages))
	for _, msg := range messages {
		chatMsgs = append(chatMsgs, ChatMessage{
			ID:        msg.ID,
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			Type:      msg.Type,
			IsMine:    msg.SenderID == info.NetID,
			CreatedAt: msg.CreatedAt,
		})
	}

	resp = ConversationDetailResponse{
		Partner: UserBrief{
			ID:        partner.ID,
			Name:      partner.Name,
			AvatarURL: partner.AvatarURL,
		},
		Messages: chatMsgs,
		Total:    total,
	}
	return resp, nil
}

// getSystemMessages 获取系统通知消息列表（partner_id=1 时调用）
func (m *MessageSvc) getSystemMessages(info GetConversationParams) (resp ConversationDetailResponse, err error) {
	var messages []model.Message
	var total int64

	query := model.DB.Model(&model.Message{}).
		Where("user_id = ?", info.NetID)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at DESC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&messages).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 标记未读系统消息为已读
	model.DB.Model(&model.Message{}).
		Where("user_id = ? AND is_read = ? AND type != ?", info.NetID, false, "chat").
		Update("is_read", true)

	chatMsgs := make([]ChatMessage, 0, len(messages))
	for _, msg := range messages {
		chatMsgs = append(chatMsgs, ChatMessage{
			ID:        msg.ID,
			SenderID:  1,
			Content:   msg.Title + "\n" + msg.Content,
			IsMine:    false,
			Type:      msg.Type,
			CreatedAt: msg.CreatedAt,
		})
	}

	resp = ConversationDetailResponse{
		Partner: UserBrief{
			ID:        1,
			Name:      "系统通知",
			AvatarURL: "", //可以添加一个系统通知的默认头像URL
		},
		Messages: chatMsgs,
		Total:    total,
	}
	return resp, nil
}

// SendChatMessage 发送聊天消息
func (m *MessageSvc) SendChatMessage(params SendChatParams) (*ChatMessage, error) {
	// 检查对方是否存在
	var partner model.User
	if err := model.DB.First(&partner, params.PartnerID).Error; err != nil {
		return nil, common.ErrNew(errors.New("对方用户不存在"), common.OpErr)
	}

	msg := &model.Message{
		NetID:    params.PartnerID,
		SenderID: params.NetID,
		Type:     "chat",
		Title:    "",
		Content:  params.Content,
		IsRead:   false,
	}

	if err := model.DB.Create(msg).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	return &ChatMessage{
		ID:        msg.ID,
		SenderID:  msg.SenderID,
		Content:   msg.Content,
		IsMine:    true,
		CreatedAt: msg.CreatedAt,
	}, nil
}
