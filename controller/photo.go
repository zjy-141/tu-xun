package controller

import (
	"fmt"
	"net/http"
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

	data, err := srv.Photo.List(params)
	if err != nil {
		fmt.Printf("controller photo list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}

// Detail 获取图片详情
func (p *Photo) Detail(c *gin.Context) {
	var params service.GetPhotoParams
	if err := c.ShouldBindUri(&params); err != nil {
		fmt.Printf("controller photo detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.CurrentUserID = SessionGet(c, "user-session").(UserSession).ID

	data, err := srv.Photo.GetByID(params)
	if err != nil {
		fmt.Printf("controller photo detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}
