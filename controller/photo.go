package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Photo struct{}

// Create 上传图片投稿（multipart/form-data）
func (ctr *Photo) Create(c *gin.Context) {
	var params service.PhotoCreateParams
	if err := c.ShouldBind(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.PhotoSvc.Create(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// List 获取题目卡片列表（可选登录）
func (ctr *Photo) List(c *gin.Context) {
	var params service.PhotoListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var userID int64
	if s, ok := SessionGet(c, "user-session").(UserSession); ok {
		userID = s.ID
	}
	resp, err := srv.PhotoSvc.List(params, userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Detail 获取题目详情（可选登录）
func (ctr *Photo) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var userID int64
	if s, ok := SessionGet(c, "user-session").(UserSession); ok {
		userID = s.ID
	}
	resp, err := srv.PhotoSvc.GetByID(id, userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// ListUser 获取当前用户投稿的题目列表
func (ctr *Photo) ListUser(c *gin.Context) {
	var params service.PhotosListUserParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.PhotoSvc.ListUser(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// DetailUser 获取我的投稿详情（仅 pending/rejected）
func (ctr *Photo) DetailUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.PhotoSvc.DetailUser(id, userID)
	if err != nil {
		c.Error(err)
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, ResponseNew(c, nil))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// PhotoComments 获取某题目下的评论列表（可选登录）
func (ctr *Photo) PhotoComments(c *gin.Context) {
	var params service.CommentListParams
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.PhotoID = id
	var userID int64
	if s, ok := SessionGet(c, "user-session").(UserSession); ok {
		userID = s.ID
	}
	resp, err := srv.PhotoSvc.PhotoComments(params, userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// PhotoSolves 获取某题目下的破解记录列表（仅 solved，可选登录）
func (ctr *Photo) PhotoSolves(c *gin.Context) {
	var params service.SolvesListParams
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.PhotoID = id
	var userID int64
	if s, ok := SessionGet(c, "user-session").(UserSession); ok {
		userID = s.ID
	}
	resp, err := srv.PhotoSvc.ListSolves(params, userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// PhotoAttemptsUser 获取某题目下当前用户的作答记录
func (ctr *Photo) PhotoAttemptsUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.AttemptSvc.ListByPhotoUser(userID, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
