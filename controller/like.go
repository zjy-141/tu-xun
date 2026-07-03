package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"

	"github.com/gin-gonic/gin"
)

type Like struct{}

// TogglePhotoLike 切换图片点赞
func (l *Like) TogglePhotoLike(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		logger.Errorf("controller like toggle_photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.LikeSvc.ToggleLike(NetID, "photo", id)
	if err != nil {
		logger.Errorf("controller like toggle_photo: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ToggleCommentLike 切换评论点赞
func (l *Like) ToggleCommentLike(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		logger.Errorf("controller like toggle_comment: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.LikeSvc.ToggleLike(NetID, "comment", id)
	if err != nil {
		logger.Errorf("controller like toggle_comment: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// GetPhotoLikeStatus 获取图片点赞状态
func (l *Like) GetPhotoLikeStatus(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		logger.Errorf("controller like status_photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.LikeSvc.GetLikeStatus(NetID, "photo", id)
	if err != nil {
		logger.Errorf("controller like status_photo: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// GetCommentLikeStatus 获取评论点赞状态
func (l *Like) GetCommentLikeStatus(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		logger.Errorf("controller like status_comment: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.LikeSvc.GetLikeStatus(NetID, "comment", id)
	if err != nil {
		logger.Errorf("controller like status_comment: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ToggleAttemptLike 切换答题记录点赞
func (l *Like) ToggleAttemptLike(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		logger.Errorf("controller like toggle_attempt: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.LikeSvc.ToggleLike(NetID, "attempt", id)
	if err != nil {
		logger.Errorf("controller like toggle_attempt: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// GetAttemptLikeStatus 获取答题记录点赞状态
func (l *Like) GetAttemptLikeStatus(c *gin.Context) {
	NetID := SessionGet(c, "user-session").(UserSession).ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		logger.Errorf("controller like status_attempt: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.LikeSvc.GetLikeStatus(NetID, "attempt", id)
	if err != nil {
		logger.Errorf("controller like status_attempt: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
