package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Good struct{}

// List 获取奖品列表
func (g *Good) List(c *gin.Context) {
	var params service.GoodListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller good list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.GoodSvc.List(params)
	if err != nil {
		logger.Errorf("service good list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Detail 获取奖品详情
func (g *Good) Detail(c *gin.Context) {
	var params service.GoodGetByIDParams
	goodId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || goodId <= 0 {
		logger.Errorf("controller good detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.GoodID = goodId

	resp, err := srv.GoodSvc.GetByID(params)
	if err != nil {
		logger.Errorf("service good detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Create 新增商品
func (g *Good) Create(c *gin.Context) {
	var params service.GoodCreateParams
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller good create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.GoodSvc.Create(params)
	if err != nil {
		logger.Errorf("service good create: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}
