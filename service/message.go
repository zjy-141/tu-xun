package service

import (
	"errors"
	"fmt"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type MessageSvc struct{}

// ==================== 通知消息 ====================

// ListMyMessages 获取当前用户的通知消息列表
func (m *MessageSvc) ListMyMessages(info ListMessageParams) (resp ListMessagesResponse, err error) {
	var messages []model.Message
	var total int64

	query := model.DB.Model(&model.Message{}).Where("user_id = ? AND type != ?", info.UserID, "chat")

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
func (m *MessageSvc) MarkAsRead(messageID int64, userID int64) error {
	result := model.DB.Model(&model.Message{}).
		Where("id = ? AND user_id = ?", messageID, userID).
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
func (m *MessageSvc) GetUnreadCount(userID int64) (resp UnreadCountResponse, err error) {
	var count int64
	if err := model.DB.Model(&model.Message{}).
		Where("user_id = ? AND is_read = ? AND type != ?", userID, false, "chat").
		Count(&count).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.Count = count
	return resp, nil
}

// SendReviewMessage 发送审核结果消息
func (m *MessageSvc) SendReviewMessage(userID int64, action string, relatedID int64, relatedType string, rejectReason string) error {
	var msgType, title, content string

	switch action {
	case "approve":
		msgType = "review_approved"
		title = fmt.Sprintf("您的%s已通过审核", relatedTypeName(relatedType))
		content = fmt.Sprintf("恭喜！您提交的%s已通过审核。", relatedTypeName(relatedType))
	case "reject":
		msgType = "review_rejected"
		title = fmt.Sprintf("您的%s未通过审核", relatedTypeName(relatedType))
		content = fmt.Sprintf("您提交的%s未通过审核。拒绝原因：%s", relatedTypeName(relatedType), rejectReason)
	default:
		return nil
	}

	msg := &model.Message{
		UserID:      userID,
		SenderID:    0,
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

func relatedTypeName(t string) string {
	switch t {
	case "photo":
		return "图片投稿"
	case "attempt":
		return "答题"
	case "comment":
		return "评论"
	default:
		return "内容"
	}
}

// ==================== 会话（微信风格聊天） ====================

// ListConversations 获取会话列表（微信首页）
func (m *MessageSvc) ListConversations(userID int64) (resp ListConversationsResponse, err error) {
	// 查询所有与我有过 chat 类型消息往来的用户
	var rows []struct {
		PartnerID int64
	}
	if err := model.DB.Raw(`
		SELECT DISTINCT partner_id FROM (
			SELECT sender_id AS partner_id FROM message WHERE user_id = ? AND type = 'chat'
			UNION
			SELECT user_id AS partner_id FROM message WHERE sender_id = ? AND type = 'chat'
		) t WHERE partner_id != 0
	`, userID, userID).Scan(&rows).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	conversations := make([]ConversationItem, 0, len(rows))
	for _, row := range rows {
		item, err := m.buildConversationItem(userID, row.PartnerID)
		if err != nil {
			continue
		}
		conversations = append(conversations, item)
	}

	// 按最新消息时间倒序
	for i := 0; i < len(conversations); i++ {
		for j := i + 1; j < len(conversations); j++ {
			if conversations[j].LastTime.After(conversations[i].LastTime) {
				conversations[i], conversations[j] = conversations[j], conversations[i]
			}
		}
	}

	resp.Conversations = conversations
	return resp, nil
}

// buildConversationItem 构建单个会话项
func (m *MessageSvc) buildConversationItem(userID, partnerID int64) (ConversationItem, error) {
	// 查对方信息
	var partner model.User
	if err := model.DB.First(&partner, partnerID).Error; err != nil {
		return ConversationItem{}, err
	}

	// 查最后一条消息
	var lastMsg model.Message
	model.DB.Where(
		"((user_id = ? AND sender_id = ?) OR (user_id = ? AND sender_id = ?)) AND type = ?",
		userID, partnerID, partnerID, userID, "chat",
	).Order("created_at DESC").First(&lastMsg)

	// 查未读数（对方发给我的未读消息）
	var unread int64
	model.DB.Model(&model.Message{}).Where(
		"user_id = ? AND sender_id = ? AND type = ? AND is_read = ?",
		userID, partnerID, "chat", false,
	).Count(&unread)

	return ConversationItem{
		PartnerID:     partner.ID,
		PartnerName:   partner.Name,
		PartnerAvatar: partner.AvatarURL,
		LastContent:   lastMsg.Content,
		LastTime:      lastMsg.CreatedAt,
		UnreadCount:   unread,
	}, nil
}

// GetConversation 获取与某用户的对话详情（微信聊天窗口）
func (m *MessageSvc) GetConversation(info GetConversationParams) (resp ConversationDetailResponse, err error) {
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
		info.UserID, info.PartnerID, "chat", false,
	).Update("is_read", true)

	// 查双方对话
	var messages []model.Message
	var total int64

	query := model.DB.Model(&model.Message{}).Where(
		"((user_id = ? AND sender_id = ?) OR (user_id = ? AND sender_id = ?)) AND type = ?",
		info.UserID, info.PartnerID, info.PartnerID, info.UserID, "chat",
	)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 按时间正序（微信风格：旧消息在上，新消息在下）
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
			IsMine:    msg.SenderID == info.UserID,
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

// SendChatMessage 发送聊天消息
func (m *MessageSvc) SendChatMessage(params SendChatParams) (*ChatMessage, error) {
	// 检查对方是否存在
	var partner model.User
	if err := model.DB.First(&partner, params.PartnerID).Error; err != nil {
		return nil, common.ErrNew(errors.New("对方用户不存在"), common.OpErr)
	}

	msg := &model.Message{
		UserID:   params.PartnerID,
		SenderID: params.UserID,
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
