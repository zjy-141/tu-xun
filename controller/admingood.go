package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type AdminGood struct{}

// ListPrizes 管理员获取所有奖品列表（已分发/未分发）
func (ag *AdminGood) List(c *gin.Context) {
	var params service.AdminListGoodsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin list prizes: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminGoodSvc.List(params)
	if err != nil {
		logger.Errorf("controller admin list prizes: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// AdminDetail 获取奖品详情
func (ag *AdminGood) Detail(c *gin.Context) {
	var params service.AdminGoodGetByIDParams
	goodId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || goodId <= 0 {
		logger.Errorf("controller good detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.GoodID = goodId

	resp, err := srv.AdminGoodSvc.GetByID(params)
	if err != nil {
		logger.Errorf("service good detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Create 新增商品
func (ag *AdminGood) Create(c *gin.Context) {
	var params service.GoodCreateParams
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller good create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminGoodSvc.Create(params)
	if err != nil {
		logger.Errorf("service good create: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// Update 更新商品
func (ag *AdminGood) Update(c *gin.Context) {
	var params service.GoodUpdateParams

	goodId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || goodId <= 0 {
		logger.Errorf("controller good detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller good create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.GoodID = goodId

	resp, err := srv.AdminGoodSvc.Update(params)
	if err != nil {
		logger.Errorf("service good create: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// Delete 获取奖品详情
func (ag *AdminGood) Delete(c *gin.Context) {
	var params service.AdminGoodGetByIDParams
	goodId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || goodId <= 0 {
		logger.Errorf("controller good detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.GoodID = goodId

	resp, err := srv.AdminGoodSvc.Delete(params)
	if err != nil {
		logger.Errorf("service good detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Status 获取奖品详情
func (ag *AdminGood) Status(c *gin.Context) {
	var params service.AdminGoodGetByIDParams
	goodId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || goodId <= 0 {
		logger.Errorf("controller good detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.GoodID = goodId

	resp, err := srv.AdminGoodSvc.Status(params)
	if err != nil {
		logger.Errorf("service good detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Delete 获取奖品详情
func (ag *AdminGood) Stock(c *gin.Context) {
	var params service.GoodUpdateStockParams
	goodId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || goodId <= 0 {
		logger.Errorf("controller good detail: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller good create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.GoodID = goodId

	resp, err := srv.AdminGoodSvc.Stock(params)
	if err != nil {
		logger.Errorf("service good detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// // ClaimPrize 标记奖品已发放
// func (ag *AdminGood) ClaimPrize(c *gin.Context) {
// 	prizeId, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil || prizeId <= 0 {
// 		logger.Errorf("controller admin claim prize: %v\n", err)
// 		c.Error(common.ErrNew(err, common.ParamErr))
// 		return
// 	}

// 	resp, err := srv.AdminSvc.ClaimPrize(prizeId)
// 	if err != nil {
// 		logger.Errorf("controller admin claim prize: %v\n", err)
// 		c.Error(err)
// 		return
// 	}

// 	c.JSON(http.StatusOK, ResponseNew(c, resp))
// }
