package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type AdminActivity struct{}

// Create 创建新活动
func (aa *AdminActivity) Create(c *gin.Context) {
	var params service.AdminActivityCreate
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller admin activity create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminActivitySvc.Create(params)
	if err != nil {
		logger.Errorf("service admin activity create: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// Update 更新活动
func (aa *AdminActivity) Update(c *gin.Context) {
	var params service.AdminActivityUpdate
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller admin activity update: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminActivitySvc.Update(params)
	if err != nil {
		logger.Errorf("service admin activity update: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Notice 发布活动公告
func (aa *AdminActivity) Notice(c *gin.Context) {
	var params service.AdminActivityNotice

	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin activity notice: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminActivitySvc.Notice(params)
	if err != nil {
		logger.Errorf("service admin activity notice: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
