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

// ListPhotos 获取题目池列表
func (a *Admin) ListPhotos(c *gin.Context) {
	var params service.AdminPhotoListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin list photos: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminSvc.ListPhotos(params)
	if err != nil {
		logger.Errorf("service admin list photos: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewPhoto 审核图片
func (a *Admin) ReviewPhoto(c *gin.Context) {
	var params service.AdminReviewPhotoParams
	photoID, err := strconv.ParseInt(c.Param("photo_id"), 10, 64)
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

// CreatePhoto 管理员新增题目（multipart/form-data）
func (a *Admin) CreatePhoto(c *gin.Context) {
	activityID, err := strconv.ParseInt(c.Param("activity_id"), 10, 64)
	if err != nil || activityID <= 0 {
		logger.Errorf("controller admin create photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	var form service.AdminPhotoUpsertForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Errorf("controller admin create photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminSvc.CreatePhoto(activityID, form)
	if err != nil {
		logger.Errorf("service admin create photo: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// UpdatePhoto 管理员编辑题目（multipart/form-data）
func (a *Admin) UpdatePhoto(c *gin.Context) {
	photoID, err := strconv.ParseInt(c.Param("photo_id"), 10, 64)
	if err != nil || photoID <= 0 {
		logger.Errorf("controller admin update photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	activityID, err := strconv.ParseInt(c.Param("activity_id"), 10, 64)
	if err != nil || activityID <= 0 {
		logger.Errorf("controller admin update photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	var form service.AdminPhotoUpsertForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Errorf("controller admin update photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminSvc.UpdatePhoto(activityID, photoID, form)
	if err != nil {
		logger.Errorf("service admin update photo: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ListAttempts 获取答题列表
func (a *Admin) ListAttempts(c *gin.Context) {
	var params service.AdminAttemptListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin list attempts: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminSvc.ListAttempts(params)
	if err != nil {
		logger.Errorf("service admin list attempts: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewAttempt 审核答题记录
func (a *Admin) ReviewAttempt(c *gin.Context) {
	var params service.AdminReviewAttemptParams
	attemptId, err := strconv.ParseInt(c.Param("photo_id"), 10, 64)
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

// ListComments 获取评论列表
func (a *Admin) ListComments(c *gin.Context) {
	var params service.AdminCommentListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin list comments: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminSvc.ListComments(params)
	if err != nil {
		logger.Errorf("service admin list comments: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewComment 审核评论
func (a *Admin) ReviewComment(c *gin.Context) {
	var params service.AdminReviewCommentParams
	commentId, err := strconv.ParseInt(c.Param("photo_id"), 10, 64)
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
		logger.Errorf("controller admin user list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminSvc.UserList(params)
	if err != nil {
		logger.Errorf("service admin user list: %v\n", err)
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

	if err := c.ShouldBind(&params); err != nil {
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

// UpdateNotification 管理员更新统一通知（multipart/form-data）
func (a *Admin) UpdateNotification(c *gin.Context) {
	var params service.UpdateNotificationRequest

	notificationID, err := strconv.ParseInt(c.Param("photo_id"), 10, 64)
	if err != nil || notificationID <= 0 {
		logger.Errorf("controller admin update notification: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.ID = notificationID

	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller admin update notification: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.MessageSvc.UpdateNotification(params)
	if err != nil {
		logger.Errorf("service admin update notification: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// DeleteNotification 管理员删除统一通知（软删除）
func (a *Admin) DeleteNotification(c *gin.Context) {
	notificationID, err := strconv.ParseInt(c.Param("photo_id"), 10, 64)
	if err != nil || notificationID <= 0 {
		logger.Errorf("controller admin delete notification: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.MessageSvc.DeleteNotification(notificationID)
	if err != nil {
		logger.Errorf("service admin delete notification: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// SetUserStatus 封禁/解封用户（仅 Level >= 3）
func (a *Admin) SetUserStatus(c *gin.Context) {
	var params service.AdminSetUserStatusParams

	userID, err := strconv.ParseInt(c.Param("photo_id"), 10, 64)
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
