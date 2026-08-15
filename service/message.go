package service

import (
	"errors"
	"fmt"
	"tu-xun/common"
	"tu-xun/model"
	"tu-xun/pkg/urlutil"
)

type MessageSvc struct{}

// targetTypeDisplayNames 将 target type 映射为中文显示名，用于自动生成消息文本
var targetTypeDisplayNames = map[string]string{
	"photo":   "题目",
	"attempt": "答题",
	"comment": "评论",
}

// ==================== 互动消息 ====================

// SendInteraction 发送互动消息（点赞/评论通知，只投递给内容所有者）。
// senderID 为操作发起者，ownerID 为内容所有者。
// 不会给自己发送通知。
func (m *MessageSvc) SendInteraction(senderID int64, targetType string, targetID int64, ownerID int64, photoID int64) {
	if senderID == ownerID || ownerID == 0 {
		return
	}
	targetName, ok := targetTypeDisplayNames[targetType]
	if !ok {
		targetName = targetType
	}
	msg := &model.InteractionMessage{
		UserID:      ownerID,
		SenderID:    senderID,
		Type:        "like",
		Content:     fmt.Sprintf("赞了你的%s", targetName),
		RelatedID:   targetID,
		RelatedType: targetType,
		PhotoID:     photoID,
		IsRead:      false,
	}
	_ = model.DB.Create(msg).Error
}

// ListInteractionMessages 获取当前用户的互动消息列表，按创建时间倒序分页。
func (m *MessageSvc) ListInteractionMessages(userID int64, params InteractionMessageListParams) (resp InteractionMessagePage, err error) {
	var msgs []model.InteractionMessage
	var total int64

	query := model.DB.Model(&model.InteractionMessage{}).Where("user_id = ?", userID)

	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.
		Order("created_at DESC, id DESC").
		Scopes(model.Paginate(params.PagerForm)).
		Preload("Sender").
		Find(&msgs).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	list := make([]InteractionMessage, 0, len(msgs))
	for _, msg := range msgs {
		list = append(list, InteractionMessage{
			ID:   msg.ID,
			Type: msg.Type,
			User: UserBrief{
				ID:       msg.Sender.ID,
				Nickname: msg.Sender.Nickname,
				Avatar:   urlutil.FullURL(msg.Sender.AvatarURL),
			},
			RelatedType: msg.RelatedType,
			RelatedID:   msg.RelatedID,
			PhotoID:     msg.PhotoID,
			Content:     msg.Content,
			IsRead:      msg.IsRead,
			CreatedAt:   &msg.CreatedAt,
		})
	}

	var unreadCount int64
	model.DB.Model(&model.InteractionMessage{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unreadCount)

	resp = InteractionMessagePage{
		Total:       total,
		List:        list,
		UnreadCount: unreadCount,
	}
	return resp, nil
}

// MarkRead 将指定互动消息标记为已读。仅当消息属于当前用户时生效。
func (m *MessageSvc) MarkRead(userID int64, id int64) error {
	result := model.DB.Model(&model.InteractionMessage{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)
	if result.Error != nil {
		return common.ErrNew(result.Error, common.SysErr)
	}
	if result.RowsAffected == 0 {
		return common.ErrNew(errors.New("消息不存在"), common.OpErr)
	}
	return nil
}

// MarkAllRead 将当前用户所有未读互动消息标记为已读。
func (m *MessageSvc) MarkAllRead(userID int64) error {
	if err := model.DB.Model(&model.InteractionMessage{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		return common.ErrNew(err, common.SysErr)
	}
	return nil
}
