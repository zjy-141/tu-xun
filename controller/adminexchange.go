package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type AdminExchange struct{}

// List 获取兑奖记录（管理端，查看所有用户）
func (ctr *AdminExchange) List(c *gin.Context) {
	var params service.AdminExchangeListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AdminExchangeSvc.List(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Verify 管理端核销/取消兑奖
func (ctr *AdminExchange) Verify(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var params service.AdminExchangeVerifyParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.ExchangeID = id
	if err := srv.AdminExchangeSvc.Verify(params); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}
