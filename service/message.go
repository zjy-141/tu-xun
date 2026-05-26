package service

import (
	"errors"
	"fmt"
	"tu-xun/common"
	"tu-xun/model"
)

type MessageSvc struct{}

// ListMyMessages 获取当前用户的消息列表
func (m *MessageSvc) ListMyMessages(info ListMessageParams) (resp ListMessagesResponse, err error) {
	var messages []model.Message
	var total int64

	query := model.DB.Model(&model.Message{}).Where("user_id = ?", info.UserID)

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

// GetUnreadCount 获取未读消息数
func (m *MessageSvc) GetUnreadCount(userID int64) (resp UnreadCountResponse, err error) {
	var count int64
	if err := model.DB.Model(&model.Message{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.Count = count
	return resp, nil
}

// SendReviewMessage 发送审核结果消息（审核通过或拒绝后调用）
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
		SenderID:    0, // 系统发送
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

// relatedTypeName 获取关联类型的显示名称
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
