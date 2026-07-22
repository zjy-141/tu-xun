package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type AdminActivity struct{}

// List 获取活动列表（分页）
func (aa *AdminActivity) List(c *gin.Context) {
	var params service.AdminActivityListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin activity list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminActivitySvc.List(params)
	if err != nil {
		logger.Errorf("service admin activity list: %v\n", err)
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Detail 获取活动详情
func (aa *AdminActivity) Detail(c *gin.Context) {
	var params service.AdminActivityGetByIDParams
	activityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || activityID <= 0 {
		logger.Errorf("controller admin activity detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.ActivityID = activityID

	resp, err := srv.AdminActivitySvc.Detail(params)
	if err != nil {
		logger.Errorf("service admin activity detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

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
