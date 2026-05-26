package controller

import (
	"net/http"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Message struct{}

// ListMyMessages 获取当前用户的消息列表
func (m *Message) ListMyMessages(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var params service.ListMessageParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, &Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	params.UserID = userID
	resp, err := srv.MessageSvc.ListMyMessages(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// MarkAsRead 标记消息为已读
func (m *Message) MarkAsRead(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var uri struct {
		ID int64 `uri:"id" binding:"min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, &Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if err := srv.MessageSvc.MarkAsRead(uri.ID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, &Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// GetUnreadCount 获取未读消息数
func (m *Message) GetUnreadCount(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := srv.MessageSvc.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
