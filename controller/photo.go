package controller

import (
	"fmt"
	"net/http"
	"tu-xun/common"

	"github.com/gin-gonic/gin"
)

type Photo struct{}

// Upload 上传图片投稿
func (p *Photo) Upload(c *gin.Context) {
	us := getCurrentUser(c)
	if us == nil {
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	locationSecret := c.PostForm("location_secret")

	if title == "" {
		c.Error(common.ErrNew(fmt.Errorf("标题不能为空"), common.ParamErr))
		return
	}
	if locationSecret == "" {
		c.Error(common.ErrNew(fmt.Errorf("具体地点不能为空"), common.ParamErr))
		return
	}

	imageFile, err := c.FormFile("image")
	if err != nil {
		c.Error(common.ErrNew(fmt.Errorf("请上传图片文件"), common.ParamErr))
		return
	}

	photo, err := srv.Photo.Create(int64(us.ID), title, description, locationSecret, imageFile)
	if err != nil {
		fmt.Printf("controller photo upload: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, photo))
}

// List 获取图片列表
func (p *Photo) List(c *gin.Context) {
	var form struct {
		common.PagerForm
		Solved *bool `form:"solved"`
	}
	if err := c.ShouldBindQuery(&form); err != nil {
		form.Page = 1
		form.Limit = 10
	}

	data, err := srv.Photo.List(form.Page, form.Limit, form.Solved)
	if err != nil {
		fmt.Printf("controller photo list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}

// Detail 获取图片详情
func (p *Photo) Detail(c *gin.Context) {
	var form common.IDUriForm
	if err := c.ShouldBindUri(&form); err != nil {
		fmt.Printf("controller photo detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	currentUserID := int64(0)
	us := getCurrentUser(c)
	if us != nil {
		currentUserID = int64(us.ID)
	}

	data, err := srv.Photo.GetByID(int64(form.ID), currentUserID)
	if err != nil {
		fmt.Printf("controller photo detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}
