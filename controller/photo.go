package controller

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Photo struct{}

// Upload 上传图片投稿
func (p *Photo) Upload(c *gin.Context) {

	var params service.CreatePhotoParams
	if err := c.ShouldBind(&params); err != nil {
		fmt.Printf("controller photo upload: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	photo, err := srv.Photo.Create(params)
	if err != nil {
		fmt.Printf("controller photo upload: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, photo))
}

// List 获取图片列表
func (p *Photo) List(c *gin.Context) {
	var params service.ListPhotoParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.Limit = 10
	}

	resp, err := srv.Photo.List(params)
	if err != nil {
		fmt.Printf("controller photo list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Detail 获取图片详情
func (p *Photo) Detail(c *gin.Context) {
	var params service.GetPhotoParams
	if err := c.ShouldBindUri(&params); err != nil {
		fmt.Printf("controller photo detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	user, ok := SessionGet(c, "user-session").(UserSession)
	if ok {
		params.CurrentUserID = user.ID
	}

	resp, err := srv.Photo.GetByID(params)
	if err != nil {
		fmt.Printf("controller photo detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Display 图片展示（流式输出原图，供 <img> 标签直接使用）
func (p *Photo) Display(c *gin.Context) {
	idStr := c.Param("id")
	photoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	imgStream, err := srv.Photo.GetImageStream(photoID)
	if err != nil {
		fmt.Printf("controller photo display: %v\n", err)
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
	idStr := c.Param("id")
	photoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	imgStream, err := srv.Photo.GetImageStream(photoID)
	if err != nil {
		fmt.Printf("controller photo download: %v\n", err)
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
