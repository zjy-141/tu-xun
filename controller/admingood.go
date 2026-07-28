package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type AdminGood struct{}

// List 管理员获取奖品列表（支持 status 和 keyword 筛选）
func (ctr *AdminGood) List(c *gin.Context) {
	var params service.AdminGoodListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminGoodSvc.List(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Create 新增商品（multipart/form-data）
func (ctr *AdminGood) Create(c *gin.Context) {
	var form service.GoodCreateParams
	if err := c.ShouldBind(&form); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminGoodSvc.Create(form)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// Update 更新商品（multipart/form-data）
func (ctr *AdminGood) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var form service.GoodUpdateParams
	if err := c.ShouldBind(&form); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	form.GoodID = id
	resp, err := srv.AdminGoodSvc.Update(form)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Delete 删除商品（有兑换记录则拒绝）
func (ctr *AdminGood) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := srv.AdminGoodSvc.Delete(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}
