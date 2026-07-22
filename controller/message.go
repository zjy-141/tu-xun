package controller

import (
	"errors"
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Message struct{}

// ==================== 统一通知 ====================

// List 获取当前用户的通知列表
func (m *Message) List(c *gin.Context) {

	var params service.NotificationListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller notification list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	// related_type 和 related_id 必须成对传
	if (params.RelatedType != "" && params.RelatedID == 0) || (params.RelatedType == "" && params.RelatedID > 0) {
		c.Error(common.ErrNew(errors.New("related_type 和 related_id 必须成对传"), common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	resp, err := srv.MessageSvc.List(params)
	if err != nil {
		logger.Errorf("service notification list: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Detail 获取通知详情
func (m *Message) Detail(c *gin.Context) {

	var params service.NotificationGetByIDParams
	NotificationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || NotificationID <= 0 {
		logger.Errorf("controller notification detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.NotificationID = NotificationID

	resp, err := srv.MessageSvc.Detail(params)
	if err != nil {
		logger.Errorf("service notification detail: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// MarkAsRead 标记通知为已读
func (m *Message) MarkAsRead(c *gin.Context) {
	var params service.NotificationReadParams
	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	notificationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || notificationID <= 0 {
		logger.Errorf("controller notification mark as read: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.NotificationID = notificationID

	err = srv.MessageSvc.MarkAsRead(params)
	if err != nil {
		logger.Errorf("service notification mark as read: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// GetUnreadCount 获取未读通知数
func (m *Message) GetUnreadCount(c *gin.Context) {
	UserID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.MessageSvc.GetUnreadCount(UserID)
	if err != nil {
		logger.Errorf("service notification get_unread_count: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// GlobalAnnouncement 获取最新一条未读且未过期的全局公告
func (m *Message) GlobalAnnouncement(c *gin.Context) {
	UserID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.MessageSvc.GetGlobalAnnouncement(UserID)
	if err != nil {
		logger.Errorf("service notification global_announcement: %v\n", err)
		c.Error(err)
		return
	}
	if resp == nil {
		c.JSON(http.StatusNoContent, nil)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
