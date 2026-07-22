package controller

import (
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
		logger.Errorf("controller admin pending photos: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
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
		c.Error(common.ErrNew(err, common.ParamErr))
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
		logger.Errorf("controller admin pending comments: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
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

// UserList 获取用户列表
func (a *Admin) UserList(c *gin.Context) {
	var params service.AdminUserListParams

	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin pending attempts: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminSvc.UserList(params)
	if err != nil {
		logger.Errorf("controller admin pending attempts: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// UpdateAdminLevel 高级管理员调整其他管理员等级
func (a *Admin) UpdateAdminLevel(c *gin.Context) {
	var params service.AdminUpdateLevelParams

	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin update level: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	session := SessionGet(c, "user-session").(UserSession)
	params.OperatorID = session.ID
	params.OperatorLevel = session.Level

	resp, err := srv.AdminSvc.UpdateAdminLevel(params)
	if err != nil {
		logger.Errorf("controller admin update level: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// SearchUsers 按关键词搜索用户（管理员用）
func (a *Admin) SearchUsers(c *gin.Context) {
	var params service.AdminSearchUsersParams

	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin search users: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminSvc.SearchUsers(params)
	if err != nil {
		logger.Errorf("service admin search users: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// CreateNotification 管理员创建统一通知
func (a *Admin) CreateNotification(c *gin.Context) {
	var params service.CreateNotificationRequest

	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin create notification: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.MessageSvc.CreateNotification(params)
	if err != nil {
		logger.Errorf("service admin create notification: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// SetUserStatus 封禁/解封用户（仅 Level >= 3）
func (a *Admin) SetUserStatus(c *gin.Context) {
	var params service.AdminSetUserStatusParams

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || userID <= 0 {
		logger.Errorf("controller admin set user status: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin set user status: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	session := SessionGet(c, "user-session").(UserSession)
	params.UserID = userID
	params.OperatorID = session.ID
	params.OperatorLevel = session.Level

	resp, err := srv.AdminSvc.SetUserStatus(params)
	if err != nil {
		logger.Errorf("service admin set user status: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
