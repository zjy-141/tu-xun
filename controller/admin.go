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

type Admin struct{}

// PendingPhotos 获取待审核图片列表
func (a *Admin) PendingPhotos(c *gin.Context) {
	var params service.AdminPendingPhotoParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin pending photos %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	params.AdminLevel = SessionGet(c, "user-session").(UserSession).Level

	resp, err := srv.AdminSvc.PendingPhotos(params)
	if err != nil {
		logger.Errorf("service admin pending photos: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewPhoto 审核图片
func (a *Admin) ReviewPhoto(c *gin.Context) {
	var params service.AdminReviewPhotoParams
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || photoID <= 0 {
		logger.Errorf("controller admin review photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin review photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.AdminLevel = SessionGet(c, "user-session").(UserSession).Level
	params.PhotoID = photoID
	resp, err := srv.AdminSvc.ReviewPhoto(params)
	if err != nil {
		logger.Errorf("service admin review photo: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// PendingAttempts 获取待审核答题记录
func (a *Admin) PendingAttempts(c *gin.Context) {
	var params service.AdminPendingAttemptParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin pending attempts: %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	params.AdminLevel = SessionGet(c, "user-session").(UserSession).Level

	resp, err := srv.AdminSvc.PendingAttempts(params)
	if err != nil {
		logger.Errorf("controller admin pending attempts: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewAttempt 审核答题记录
func (a *Admin) ReviewAttempt(c *gin.Context) {
	var params service.AdminReviewAttemptParams
	attemptId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || attemptId <= 0 {
		logger.Errorf("controller admin review attempt: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin review attempt: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.AdminLevel = SessionGet(c, "user-session").(UserSession).Level
	params.AttemptID = attemptId

	resp, err := srv.AdminSvc.ReviewAttempt(params)
	if err != nil {
		logger.Errorf("controller admin review attempt: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// PendingComments 获取待审核评论
func (a *Admin) PendingComments(c *gin.Context) {
	var params service.AdminPendingCommentParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.Limit = 10
	}
	params.AdminLevel = SessionGet(c, "user-session").(UserSession).Level

	resp, err := srv.AdminSvc.PendingComments(params)
	if err != nil {
		logger.Errorf("controller admin pending comments: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewComment 审核评论
func (a *Admin) ReviewComment(c *gin.Context) {
	var params service.AdminReviewCommentParams
	commentId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || commentId <= 0 {
		logger.Errorf("controller admin review comment: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin review comment: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.CommentID = commentId

	resp, err := srv.AdminSvc.ReviewComment(params)
	if err != nil {
		logger.Errorf("controller admin review comment: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Announcement 全服公告
func (a *Admin) Announcement(c *gin.Context) {
	var params service.AdminAnnouncement

	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin review comment: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminSvc.Announcement(params)
	if err != nil {
		logger.Errorf("controller admin review comment: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// UpdateAdminLevel 高级管理员调整其他管理员等级
func (a *Admin) UpdateAdminLevel(c *gin.Context) {
	var params service.AdminUpdateLevelParams
	UserID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || UserID <= 0 {
		logger.Errorf("controller admin update level: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin update level: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	sess := SessionGet(c, "user-session").(UserSession)
	params.UserID = UserID
	params.OperatorID = sess.ID
	params.OperatorLevel = sess.Level

	resp, err := srv.AdminSvc.UpdateAdminLevel(params)
	if err != nil {
		logger.Errorf("controller admin update level: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
