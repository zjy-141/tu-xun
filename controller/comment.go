package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Comment struct{}

// Create 发表评论
func (co *Comment) Create(c *gin.Context) {
	var params service.CreateCommentParams
	if err := c.ShouldBindUri(&params); err != nil {
		logger.Errorf("controller comment create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller comment create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	resp, err := srv.Comment.Create(params)
	if err != nil {
		logger.Errorf("controller comment create: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// UserComments 获取某用户的所有评论（个人主页用）
func (co *Comment) UserComments(c *gin.Context) {
	var params service.ListUserCommentsParams
	if err := c.ShouldBindUri(&params); err != nil {
		logger.Errorf("controller comment user comments: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller comment user comments: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.Comment.ListByUser(params)
	if err != nil {
		logger.Errorf("controller comment user comments: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
