package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type AdminExchange struct{}

// List 获取兑奖记录（管理端，查看所有用户）
func (ae *AdminExchange) List(c *gin.Context) {
	var params service.AdminExchangeListParams

	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller admin exchange list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminExchangeSvc.List(params)
	if err != nil {
		logger.Errorf("service admin exchange list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Verify 管理端核销/取消兑奖
func (ae *AdminExchange) Verify(c *gin.Context) {
	var params service.AdminExchangeVerifyParams
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller admin exchange verify: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	resp, err := srv.AdminExchangeSvc.Verify(params)
	if err != nil {
		logger.Errorf("service admin exchange verify: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
