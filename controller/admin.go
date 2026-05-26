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
	var params service.PendingPhotoParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin pending photos %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	params.AdminLevel = SessionGet(c, "user-session").(UserSession).Level

	resp, err := srv.Admin.PendingPhotos(params)
	if err != nil {
		logger.Errorf("controller admin pending photos: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewPhoto 审核图片
func (a *Admin) ReviewPhoto(c *gin.Context) {
	var params service.ReviewPhotoParams
	photoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || photoId <= 0 {
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
	params.PhotoID = photoId
	resp, err := srv.Admin.ReviewPhoto(params)
	if err != nil {
		logger.Errorf("controller admin review photo: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// PendingAttempts 获取待审核答题记录
func (a *Admin) PendingAttempts(c *gin.Context) {
	var params service.PendingAttemptParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin pending attempts: %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	params.AdminLevel = SessionGet(c, "user-session").(UserSession).Level

	resp, err := srv.Admin.PendingAttempts(params)
	if err != nil {
		logger.Errorf("controller admin pending attempts: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewAttempt 审核答题记录
func (a *Admin) ReviewAttempt(c *gin.Context) {
	var params service.ReviewAttemptParams
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

	resp, err := srv.Admin.ReviewAttempt(params)
	if err != nil {
		logger.Errorf("controller admin review attempt: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// PendingComments 获取待审核评论
func (a *Admin) PendingComments(c *gin.Context) {
	var params common.PagerForm
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.Limit = 10
	}

	resp, err := srv.Admin.PendingComments(params)
	if err != nil {
		logger.Errorf("controller admin pending comments: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewComment 审核评论
func (a *Admin) ReviewComment(c *gin.Context) {
	var params service.ReviewCommentParams
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

	resp, err := srv.Admin.ReviewComment(params)
	if err != nil {
		logger.Errorf("controller admin review comment: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ClaimPrize 标记奖品已发放
func (a *Admin) ClaimPrize(c *gin.Context) {
	prizeId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || prizeId <= 0 {
		logger.Errorf("controller admin claim prize: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.Admin.ClaimPrize(prizeId)
	if err != nil {
		logger.Errorf("controller admin claim prize: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// UpdateAdminLevel 高级管理员调整其他管理员等级
func (a *Admin) UpdateAdminLevel(c *gin.Context) {
	var params service.UpdateAdminLevelParams
	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || userId <= 0 {
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
	params.UserID = userId
	params.OperatorID = sess.ID
	params.OperatorLevel = sess.Level

	resp, err := srv.Admin.UpdateAdminLevel(params)
	if err != nil {
		logger.Errorf("controller admin update level: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
