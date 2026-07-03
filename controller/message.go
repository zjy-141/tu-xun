package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Message struct{}

// ==================== 通知消息 ====================

// ListMyMessages 获取当前用户的通知消息列表
func (m *Message) ListMyMessages(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	var params service.ListMessageParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller message list_my_messages: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.NetID = NetID
	resp, err := srv.MessageSvc.ListMyMessages(params)
	if err != nil {
		logger.Errorf("controller message list_my_messages: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// MarkAsRead 标记消息为已读
func (m *Message) MarkAsRead(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		logger.Errorf("controller message mark_as_read: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	err = srv.MessageSvc.MarkAsRead(id, NetID)
	if err != nil {
		logger.Errorf("controller message mark_as_read: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// GetUnreadCount 获取未读通知数
func (m *Message) GetUnreadCount(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.MessageSvc.GetUnreadCount(NetID)
	if err != nil {
		logger.Errorf("controller message get_unread_count: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ==================== 会话（微信风格聊天） ====================

// ListConversations 获取会话列表（微信首页）
func (m *Message) ListConversations(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.MessageSvc.ListConversations(NetID)
	if err != nil {
		logger.Errorf("controller message list_conversations: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// GetConversation 获取与某用户的对话详情（微信聊天窗口）
func (m *Message) GetConversation(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	var params service.GetConversationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller message get_conversation: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	partnerId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || partnerId <= 0 {
		logger.Errorf("controller message get_conversation: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.PartnerID = partnerId
	params.NetID = NetID
	resp, err := srv.MessageSvc.GetConversation(params)
	if err != nil {
		logger.Errorf("controller message get_conversation: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// SendChatMessage 发送聊天消息
func (m *Message) SendChatMessage(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	var params service.SendChatParams
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller message send_chat: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	partnerId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || partnerId <= 0 {
		logger.Errorf("controller message send_chat: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.PartnerID = partnerId
	params.NetID = NetID
	msg, err := srv.MessageSvc.SendChatMessage(params)
	if err != nil {
		logger.Errorf("controller message send_chat: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, ResponseNew(c, msg))
}
