package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Message struct{}

// ListInteractionMessages 获取当前用户的互动消息列表
func (ctr *Message) ListInteractionMessages(c *gin.Context) {
	var params service.InteractionMessageListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.MessageSvc.ListInteractionMessages(userID, params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// MarkInteractionRead 将指定互动消息标记为已读
func (ctr *Message) MarkInteractionRead(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userID := SessionGet(c, "user-session").(UserSession).ID
	if err := srv.MessageSvc.MarkRead(userID, id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// MarkAllInteractionRead 将所有互动消息标记为已读
func (ctr *Message) MarkAllInteractionRead(c *gin.Context) {
	userID := SessionGet(c, "user-session").(UserSession).ID
	if err := srv.MessageSvc.MarkAllRead(userID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}
