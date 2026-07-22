package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Like struct{}

// SetPhotoLike 幂等设置图片点赞状态（PUT）
func (l *Like) SetPhotoLike(c *gin.Context) {
	var params service.LikeTarget

	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		logger.Errorf("controller like set photo like: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller like set photo like: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.TargetType = "photo"
	params.TargetID = targetID

	resp, err := srv.LikeSvc.SetLike(params)
	if err != nil {
		logger.Errorf("service like set photo like: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// SetCommentLike 幂等设置评论点赞状态（PUT）
func (l *Like) SetCommentLike(c *gin.Context) {
	var params service.LikeTarget

	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		logger.Errorf("controller like set comment like: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller like set comment like: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.TargetType = "comment"
	params.TargetID = targetID

	resp, err := srv.LikeSvc.SetLike(params)
	if err != nil {
		logger.Errorf("service like set comment like: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// SetAttemptLike 幂等设置答题记录点赞状态（PUT）
func (l *Like) SetAttemptLike(c *gin.Context) {
	var params service.LikeTarget

	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		logger.Errorf("controller like set attempt like: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller like set attempt like: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.TargetType = "attempt"
	params.TargetID = targetID

	resp, err := srv.LikeSvc.SetLike(params)
	if err != nil {
		logger.Errorf("service like set attempt like: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// GetPhotoLikeStatus 获取图片点赞状态
func (l *Like) GetPhotoLikeStatus(c *gin.Context) {
	var params service.LikeTarget

	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		logger.Errorf("controller like get photo like status: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.TargetType = "photo"
	params.TargetID = targetID

	resp, err := srv.LikeSvc.GetLikeStatus(params)
	if err != nil {
		logger.Errorf("service like get photo like status: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// GetCommentLikeStatus 获取评论点赞状态
func (l *Like) GetCommentLikeStatus(c *gin.Context) {
	var params service.LikeTarget

	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		logger.Errorf("controller like get comment like status: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.TargetType = "comment"
	params.TargetID = targetID

	resp, err := srv.LikeSvc.GetLikeStatus(params)
	if err != nil {
		logger.Errorf("service like get comment like status: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// GetAttemptLikeStatus 获取答题记录点赞状态
func (l *Like) GetAttemptLikeStatus(c *gin.Context) {
	var params service.LikeTarget

	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		logger.Errorf("controller like get attempt like status: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.TargetType = "attempt"
	params.TargetID = targetID

	resp, err := srv.LikeSvc.GetLikeStatus(params)
	if err != nil {
		logger.Errorf("service like get attempt like status: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
