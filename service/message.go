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

// List 获取当前用户的通知消息列表
func (m *MessageSvc) List(info MessageListParams) (resp MessageForms, err error) {
	var messages []model.Message
	var total int64

	query := model.DB.Model(&model.Message{}).Where("user_id in (?,?) AND type != ?", info.UserID, -1, "chat")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at DESC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&messages).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Messages = make([]MessageForm, 0, len(messages))
	for _, msg := range messages {
		resp.Messages = append(resp.Messages, MessageForm{
			ID:        msg.ID,
			SenderID:  msg.SenderID,
			Title:     msg.Title,
			IsRead:    msg.IsRead,
			CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}

// Detail 获取通知详情
func (m *MessageSvc) Detail(info MessageGetByIDParams) (resp MessageDetail, err error) {
	var message model.Message
	if err := model.DB.First(&message, info.MessageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("消息不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = MessageDetail{
		ID:          message.ID,
		UserID:      message.UserID,
		SenderID:    message.SenderID,
		Type:        message.Type,
		Title:       message.Title,
		Content:     message.Content,
		RelatedID:   message.RelatedID,
		RelatedType: message.RelatedType,
		IsRead:      message.IsRead,
		CreatedAt:   message.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

// MarkAsRead 标记消息为已读
func (m *MessageSvc) MarkAsRead(info MessageReadedParams) (err error) {
	result := model.DB.Model(&model.Message{}).
		Where("id = ? AND user_id = ?", info.MessageID, info.UserID).
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
func (m *MessageSvc) GetUnreadCount(userID int64) (resp MessageUnreadCount, err error) {
	var count int64
	if err := model.DB.Model(&model.Message{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.Count = count
	return resp, nil
}

func (m *MessageSvc) Notice(info MessageNoticeParams) (resp MessageNoticeForms, err error) {
	var messages []model.Message
	var total int64

	query := model.DB.Model(&model.Message{}).Where("related_id = ? AND related_type = ? AND type = ?", info.ActivityID, "activity", "notice")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at DESC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&messages).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.MessageNotices = make([]MessageNoticeForm, 0, len(messages))
	for _, msg := range messages {
		resp.MessageNotices = append(resp.MessageNotices, MessageNoticeForm{
			Title:       msg.Title,
			Content:     msg.Content,
			RelatedID:   msg.RelatedID,
			RelatedType: msg.RelatedType,
			CreatedAt:   msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}

// FeedBack 发送反馈
func (m *MessageSvc) FeedBack(info MessageFeadBack) (resp ResponseIS, err error) {

	msg := &model.Message{
		UserID:   1, // 发送给系统
		SenderID: info.UserID,
		Type:     "feedback",
		Title:    info.Title,
		Content:  info.Content,
		IsRead:   false,
	}

	if err := model.DB.Create(msg).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp = ResponseIS{
		ID:     msg.ID,
		Status: "unRead",
	}
	return resp, nil
}

// SendLikeNotification 发送点赞通知
func (m *MessageSvc) SendLikeNotification(likerID int64, targetType string, targetID int64, ownerID int64) {
	if likerID == ownerID {
		return // 不给自己发通知
	}
	msg := &model.Message{
		UserID:      ownerID,
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

var relatedTypeNames = map[string]string{
	"photo":   "图片投稿",
	"attempt": "答题",
	"comment": "评论",
}

// // SendAttemptNotification 发送答题通知
// func (m *MessageSvc) SendAttemptNotification(submitterID int64, photoID int64, ownerID int64) {
// 	if submitterID == ownerID {
// 		return // 不给自己发通知
// 	}
// 	msg := &model.Message{
// 		UserID:      ownerID,
// 		SenderID:    1,
// 		Type:        "attempt",
// 		Title:       "有人挑战了你的图片",
// 		Content:     "有人提交了新的答题，等待管理员审核。",
// 		RelatedID:   photoID,
// 		RelatedType: "photo",
// 		IsRead:      false,
// 	}
// 	_ = model.DB.Create(msg).Error
// }

// // ==================== 会话（微信风格聊天） ====================

// // ListConversations 获取会话列表（微信首页：系统通知 + 用户聊天）
// func (m *MessageSvc) ListConversations(userID int64) (resp ListConversationsResponse, err error) {
// 	var rows []struct {
// 		PartnerID int64
// 	}
// 	if err := model.DB.Raw(`
// 		SELECT DISTINCT partner_id FROM (
// 			SELECT 1 AS partner_id FROM message WHERE user_id = ? AND type != 'chat'
// 			UNION
// 			SELECT sender_id AS partner_id FROM message WHERE user_id = ? AND type = 'chat'
// 			UNION
// 			SELECT user_id AS partner_id FROM message WHERE sender_id = ? AND type = 'chat'
// 		) t
// 	`, userID, userID, userID).Scan(&rows).Error; err != nil {
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	conversations := make([]ConversationItem, 0, len(rows))
// 	for _, row := range rows {
// 		var item ConversationItem
// 		var itemErr error
// 		if row.PartnerID == 1 {
// 			item, itemErr = m.buildSystemConversation(userID)
// 		} else {
// 			item, itemErr = m.buildConversationItem(userID, row.PartnerID)
// 		}
// 		if itemErr != nil {
// 			logger.Errorf("构建会话项失败: userID=%d, partnerID=%d, error=%v\n", userID, row.PartnerID, itemErr)
// 			continue
// 		}
// 		conversations = append(conversations, item)
// 	}

// 	// 按最新消息时间倒序
// 	sortConversations(conversations)
// 	resp.Conversations = conversations
// 	return resp, nil
// }

// // buildSystemConversation 构建系统通知虚拟会话
// func (m *MessageSvc) buildSystemConversation(userID int64) (item ConversationItem, err error) {
// 	// 查最后一条消息
// 	var lastMsg model.Message
// 	err = model.DB.Where("user_id = ? AND type != ?", userID, "chat").
// 		Order("created_at DESC").First(&lastMsg).Error
// 	if err != nil {
// 		return item, common.ErrNew(err, common.SysErr)
// 	}

// 	// 查未读数（对方发给我的未读消息）
// 	var unread int64
// 	model.DB.Model(&model.Message{}).
// 		Where("user_id = ? AND is_read = ? AND type != ?", userID, false, "chat").
// 		Count(&unread)

// 	item = ConversationItem{
// 		PartnerID:     1,
// 		PartnerName:   "系统通知",
// 		PartnerAvatar: "",
// 		LastContent:   lastMsg.Title,
// 		LastTime:      lastMsg.CreatedAt,
// 		UnreadCount:   unread,
// 	}
// 	return item, nil
// }

// // sortConversations 按 LastTime 降序冒泡排序
// func sortConversations(items []ConversationItem) {
// 	for i := range items {
// 		for j := i + 1; j < len(items); j++ {
// 			if items[j].LastTime.After(items[i].LastTime) {
// 				items[i], items[j] = items[j], items[i]
// 			}
// 		}
// 	}
// }

// // buildConversationItem 构建单个会话项
// func (m *MessageSvc) buildConversationItem(userID, partnerID int64) (item ConversationItem, err error) {
// 	// 查对方信息
// 	var partner model.User
// 	if err := model.DB.First(&partner, partnerID).Error; err != nil {
// 		return item, common.ErrNew(err, common.OpErr)
// 	}

// 	// 查最后一条消息
// 	var lastMsg model.Message
// 	model.DB.Where(
// 		"((user_id = ? AND sender_id = ?) OR (user_id = ? AND sender_id = ?)) AND type = ?",
// 		userID, partnerID, partnerID, userID, "chat",
// 	).Order("created_at DESC").First(&lastMsg)

// 	// 查未读数（对方发给我的未读消息）
// 	var unread int64
// 	model.DB.Model(&model.Message{}).Where(
// 		"user_id = ? AND sender_id = ? AND type = ? AND is_read = ?",
// 		userID, partnerID, "chat", false,
// 	).Count(&unread)

// 	item = ConversationItem{
// 		PartnerID:     partner.ID,
// 		PartnerName:   partner.Name,
// 		PartnerAvatar: partner.AvatarURL,
// 		LastContent:   lastMsg.Content,
// 		LastTime:      lastMsg.CreatedAt,
// 		UnreadCount:   unread,
// 	}
// 	return item, nil
// }

// // GetConversation 获取对话详情（微信聊天窗口 / 系统通知列表）
// func (m *MessageSvc) GetConversation(info GetConversationParams) (resp ConversationDetailResponse, err error) {
// 	// --- partner_id=1 → 系统通知 ---
// 	if info.PartnerID == 1 {
// 		return m.getSystemMessages(info)
// 	}

// 	// --- partner_id>0 → 用户聊天 ---
// 	// 查对方信息
// 	var partner model.User
// 	if err := model.DB.First(&partner, info.PartnerID).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return resp, common.ErrNew(errors.New("用户不存在"), common.OpErr)
// 		}
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	// 标记对方发来的未读消息为已读
// 	model.DB.Model(&model.Message{}).Where(
// 		"user_id = ? AND sender_id = ? AND type = ? AND is_read = ?",
// 		info.UserID, info.PartnerID, "chat", false,
// 	).Update("is_read", true)

// 	// 查双方对话
// 	var messages []model.Message
// 	var total int64

// 	query := model.DB.Model(&model.Message{}).Where(
// 		"((user_id = ? AND sender_id = ?) OR (user_id = ? AND sender_id = ?)) AND type = ?",
// 		info.UserID, info.PartnerID, info.PartnerID, info.UserID, "chat",
// 	)

// 	if err := query.Count(&total).Error; err != nil {
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	if err := query.Order("created_at ASC").
// 		Scopes(model.Paginate(info.PagerForm)).
// 		Find(&messages).Error; err != nil {
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	chatMsgs := make([]ChatMessage, 0, len(messages))
// 	for _, msg := range messages {
// 		chatMsgs = append(chatMsgs, ChatMessage{
// 			ID:        msg.ID,
// 			SenderID:  msg.SenderID,
// 			Content:   msg.Content,
// 			Type:      msg.Type,
// 			IsMine:    msg.SenderID == info.UserID,
// 			CreatedAt: msg.CreatedAt,
// 		})
// 	}

// 	resp = ConversationDetailResponse{
// 		Partner: UserBrief{
// 			ID:        partner.ID,
// 			Name:      partner.Name,
// 			AvatarURL: partner.AvatarURL,
// 		},
// 		Messages: chatMsgs,
// 		Total:    total,
// 	}
// 	return resp, nil
// }

// // getSystemMessages 获取系统通知消息列表（partner_id=1 时调用）
// func (m *MessageSvc) getSystemMessages(info GetConversationParams) (resp ConversationDetailResponse, err error) {
// 	var messages []model.Message
// 	var total int64

// 	query := model.DB.Model(&model.Message{}).
// 		Where("user_id = ?", info.UserID)

// 	if err := query.Count(&total).Error; err != nil {
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	if err := query.Order("created_at DESC").
// 		Scopes(model.Paginate(info.PagerForm)).
// 		Find(&messages).Error; err != nil {
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	// 标记未读系统消息为已读
// 	model.DB.Model(&model.Message{}).
// 		Where("user_id = ? AND is_read = ? AND type != ?", info.UserID, false, "chat").
// 		Update("is_read", true)

// 	chatMsgs := make([]ChatMessage, 0, len(messages))
// 	for _, msg := range messages {
// 		chatMsgs = append(chatMsgs, ChatMessage{
// 			ID:        msg.ID,
// 			SenderID:  1,
// 			Content:   msg.Title + "\n" + msg.Content,
// 			IsMine:    false,
// 			Type:      msg.Type,
// 			CreatedAt: msg.CreatedAt,
// 		})
// 	}

// 	resp = ConversationDetailResponse{
// 		Partner: UserBrief{
// 			ID:        1,
// 			Name:      "系统通知",
// 			AvatarURL: "", //可以添加一个系统通知的默认头像URL
// 		},
// 		Messages: chatMsgs,
// 		Total:    total,
// 	}
// 	return resp, nil
// }

// // SendChatMessage 发送聊天消息
// func (m *MessageSvc) SendChatMessage(params SendChatParams) (*ChatMessage, error) {
// 	// 检查对方是否存在
// 	var partner model.User
// 	if err := model.DB.First(&partner, params.PartnerID).Error; err != nil {
// 		return nil, common.ErrNew(errors.New("对方用户不存在"), common.OpErr)
// 	}

// 	msg := &model.Message{
// 		UserID:   params.PartnerID,
// 		SenderID: params.UserID,
// 		Type:     "chat",
// 		Title:    "",
// 		Content:  params.Content,
// 		IsRead:   false,
// 	}

// 	if err := model.DB.Create(msg).Error; err != nil {
// 		return nil, common.ErrNew(err, common.SysErr)
// 	}

// 	return &ChatMessage{
// 		ID:        msg.ID,
// 		SenderID:  msg.SenderID,
// 		Content:   msg.Content,
// 		IsMine:    true,
// 		CreatedAt: msg.CreatedAt,
// 	}, nil
// }
