package controller

import (
	"fmt"
	"net/http"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Story struct{}

// Create 发布故事
func (s *Story) Create(c *gin.Context) {

	var params service.CreateStoryParams
	if err := c.ShouldBindUri(&params); err != nil {
		fmt.Printf("controller story create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		fmt.Printf("controller story create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	story, err := srv.Story.Create(params)
	if err != nil {
		fmt.Printf("controller story create: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, story))
}

// UploadMedia 上传故事媒体文件
func (s *Story) UploadMedia(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		fmt.Printf("controller story uploadMedia: %v\n", err)
		c.Error(common.ErrNew(fmt.Errorf("请上传媒体文件"), common.ParamErr))
		return
	}

	mediaURL, err := srv.Story.UploadMedia(file)
	if err != nil {
		fmt.Printf("controller story uploadMedia: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, gin.H{
		"media_url": mediaURL,
	}))
}

// ListByPhoto 获取图片下的故事列表
func (s *Story) ListByPhoto(c *gin.Context) {
	var params service.ListStoryByPhotoParams
	if err := c.ShouldBindUri(&params); err != nil {
		fmt.Printf("controller story list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	data, err := srv.Story.ListByPhoto(params)
	if err != nil {
		fmt.Printf("controller story list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}
