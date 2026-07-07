package controller

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Photo struct{}

// Create 上传图片投稿
func (p *Photo) Create(c *gin.Context) {
	var params service.PhotoCreateParams
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller photo create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.PhotoSvc.Create(params)
	if err != nil {
		logger.Errorf("service photo create: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// List 获取图片列表
func (p *Photo) List(c *gin.Context) {
	var params service.PhotoListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller photo list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.PhotoSvc.List(params)
	if err != nil {
		logger.Errorf("service photo list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Detail 获取图片详情
func (p *Photo) Detail(c *gin.Context) {
	var params service.PhotoGetByIDParams
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || photoID <= 0 {
		logger.Errorf("controller photo detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.PhotoID = photoID

	resp, err := srv.PhotoSvc.GetByID(params)
	if err != nil {
		logger.Errorf("service photo detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// GetImageStream 获取图片流（流式输出原图，供 <img> 标签直接使用）
func (p *Photo) GetImageStream(c *gin.Context) {
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || photoID <= 0 {
		logger.Errorf("controller photo get image stream: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	imgStream, err := srv.PhotoSvc.GetImageStream(photoID)
	if err != nil {
		logger.Errorf("service photo get image stream: %v\n", err)
		c.Error(err)
		return
	}
	defer imgStream.Reader.Close()

	// 设置缓存头（浏览器缓存 1 小时）
	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("Content-Type", imgStream.ContentType)

	if imgStream.Size > 0 {
		c.Header("Content-Length", strconv.FormatInt(imgStream.Size, 10))
	}

	c.Status(http.StatusOK)
	io.Copy(c.Writer, imgStream.Reader)
}

// Download 图片下载（强制浏览器下载）
func (p *Photo) Download(c *gin.Context) {
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || photoID <= 0 {
		logger.Errorf("controller photo download: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	imgStream, err := srv.PhotoSvc.GetImageStream(photoID)
	if err != nil {
		logger.Errorf("service photo download: %v\n", err)
		c.Error(err)
		return
	}
	defer imgStream.Reader.Close()

	// 设置下载头
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, imgStream.Filename))
	c.Header("Content-Type", imgStream.ContentType)

	if imgStream.Size > 0 {
		c.Header("Content-Length", strconv.FormatInt(imgStream.Size, 10))
	}

	c.Status(http.StatusOK)
	io.Copy(c.Writer, imgStream.Reader)
}

// UserPhotos 获取某用户投稿的图片列表（个人主页用）
func (p *Photo) UserPhotos(c *gin.Context) {
	var params service.PhotosListUserParams

	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller photo user photos: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	resp, err := srv.PhotoSvc.ListByUser(params)
	if err != nil {
		logger.Errorf("controller photo user photos: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// PhotoComments 获取某图片下的评论列表
func (p *Photo) PhotoComments(c *gin.Context) {
	var params service.PhotoCommentsListParams
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || photoID <= 0 {
		logger.Errorf("controller photo comments: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller photo comments: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.PhotoID = photoID

	resp, err := srv.CommentSvc.ListByPhoto(params)
	if err != nil {
		logger.Errorf("service photo comments: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// PhotoAttempts 获取某图片下的答题记录列表
func (p *Photo) PhotoAttempts(c *gin.Context) {
	var params service.PhotoAttemptsListParams
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || photoID <= 0 {
		logger.Errorf("controller photo attempts: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller photo attempts: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.PhotoID = photoID

	resp, err := srv.AttemptSvc.ListByPhoto(params)
	if err != nil {
		logger.Errorf("service photo attempts: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
