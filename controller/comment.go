package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Comment struct{}

// Create 发表评论
func (co *Comment) Create(c *gin.Context) {
	var params service.CreateCommentParams
	photoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || photoId <= 0 {
		logger.Errorf("controller comment create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller comment create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.PhotoID = photoId
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
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || userId <= 0 {
		logger.Errorf("controller comment user comments: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller comment user comments: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = userId

	resp, err := srv.Comment.ListByUser(params)
	if err != nil {
		logger.Errorf("controller comment user comments: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
