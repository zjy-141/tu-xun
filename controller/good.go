package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Good struct{}

// List 获取上架奖品列表
func (ctr *Good) List(c *gin.Context) {
	var params service.GoodListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.GoodSvc.List(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
