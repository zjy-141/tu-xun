package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Admin struct{}

// ==================== Photo Management ====================

// ListPhotos 获取题目池列表
func (ctr *Admin) ListPhotos(c *gin.Context) {
	var params service.AdminPhotoListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminSvc.ListPhotos(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewPhoto 审核图片
func (ctr *Admin) ReviewPhoto(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var params service.AdminReviewPhotoParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.PhotoID = id
	params.AdminLevel = SessionGet(c, "user-session").(UserSession).Level
	resp, err := srv.AdminSvc.ReviewPhoto(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// CreatePhoto 管理员新增题目（multipart/form-data）
func (ctr *Admin) CreatePhoto(c *gin.Context) {
	var form service.AdminPhotoCreateForm
	if err := c.ShouldBind(&form); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminSvc.CreatePhoto(form)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// UpdatePhoto 管理员编辑题目（multipart/form-data）
func (ctr *Admin) UpdatePhoto(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var form service.AdminPhotoUpdateForm
	if err := c.ShouldBind(&form); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	form.PhotoID = id
	resp, err := srv.AdminSvc.UpdatePhoto(form)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ==================== Attempt Management ====================

// ListAttempts 获取答题列表
func (ctr *Admin) ListAttempts(c *gin.Context) {
	var params service.AdminAttemptListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminSvc.ListAttempts(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewAttempt 审核答题记录
func (ctr *Admin) ReviewAttempt(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var params service.AdminReviewAttemptParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.AttemptID = id
	params.AdminLevel = SessionGet(c, "user-session").(UserSession).Level
	resp, err := srv.AdminSvc.ReviewAttempt(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ==================== Comment Management ====================

// ListComments 获取评论列表
func (ctr *Admin) ListComments(c *gin.Context) {
	var params service.AdminCommentListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminSvc.ListComments(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ReviewComment 审核评论
func (ctr *Admin) ReviewComment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var params service.AdminReviewCommentParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.CommentID = id
	resp, err := srv.AdminSvc.ReviewComment(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ==================== User Management ====================

// ListUsers 获取用户列表（Level 3 专用）
func (ctr *Admin) ListUsers(c *gin.Context) {
	var params service.AdminUserListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminSvc.ListUsers(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// UpdateAdminLevel 高级管理员调整其他管理员等级
func (ctr *Admin) UpdateAdminLevel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var params service.AdminUpdateLevelParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	session := SessionGet(c, "user-session").(UserSession)
	params.UserID = id
	params.OperatorID = session.ID
	params.OperatorLevel = session.Level
	if err := srv.AdminSvc.UpdateAdminLevel(params); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// SetUserStatus 封禁/解封用户（仅 Level >= 3）
func (ctr *Admin) SetUserStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var params service.AdminSetUserStatusParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	session := SessionGet(c, "user-session").(UserSession)
	params.UserID = id
	params.OperatorID = session.ID
	params.OperatorLevel = session.Level
	if err := srv.AdminSvc.SetUserStatus(params); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// ==================== Stats ====================

// GetStats 获取管理端工作台统计数据
func (ctr *Admin) GetStats(c *gin.Context) {
	resp, err := srv.StatsSvc.GetStats()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
