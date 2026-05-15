package controller

import (
	"fmt"
	"net/http"
	"tu-xun/common"

	"github.com/gin-gonic/gin"
)

type Story struct{}

// Create 发布故事
func (s *Story) Create(c *gin.Context) {
	us := getCurrentUser(c)
	if us == nil {
		return
	}

	var uriForm common.IDUriForm
	if err := c.ShouldBindUri(&uriForm); err != nil {
		fmt.Printf("controller story create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	var body struct {
		Content  string `json:"content" binding:"required"`
		MediaURL string `json:"media_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Printf("controller story create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	story, err := srv.Story.Create(int64(uriForm.ID), int64(us.ID), body.Content, body.MediaURL)
	if err != nil {
		fmt.Printf("controller story create: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, story))
}

// ListByPhoto 获取图片下的故事列表
func (s *Story) ListByPhoto(c *gin.Context) {
	var uriForm common.IDUriForm
	if err := c.ShouldBindUri(&uriForm); err != nil {
		fmt.Printf("controller story list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	data, err := srv.Story.ListByPhoto(int64(uriForm.ID))
	if err != nil {
		fmt.Printf("controller story list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}
