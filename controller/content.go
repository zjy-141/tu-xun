package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type ContentBlock struct{}

// Get 按 key 获取内容位
func (ctr *ContentBlock) Get(c *gin.Context) {
	key := c.Param("key")
	resp, err := srv.ContentBlockSvc.GetByKey(key)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// AdminUpdate 管理端更新内容位
func (ctr *ContentBlock) AdminUpdate(c *gin.Context) {
	key := c.Param("key")
	var params service.UpdateContentRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := srv.ContentBlockSvc.AdminUpdate(key, params); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}
