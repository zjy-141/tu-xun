package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type AdminActivity struct{}

// List 获取活动列表（分页，支持状态和关键词筛选）
func (ctr *AdminActivity) List(c *gin.Context) {
	var params service.AdminActivityListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminActivitySvc.List(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Create 创建新活动（multipart/form-data）
func (ctr *AdminActivity) Create(c *gin.Context) {
	var form service.AdminActivityCreate
	if err := c.ShouldBind(&form); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminActivitySvc.Create(form)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// Update 更新活动（multipart/form-data）
func (ctr *AdminActivity) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var form service.AdminActivityUpdate
	if err := c.ShouldBind(&form); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	form.ActivityID = id
	resp, err := srv.AdminActivitySvc.Update(form)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Delete 删除活动
func (ctr *AdminActivity) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminActivitySvc.Delete(id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
