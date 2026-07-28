package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Activity struct{}

// List 获取活动卡片列表（含状态筛选和关键词搜索）
func (ctr *Activity) List(c *gin.Context) {
	var params service.ActivityListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.ActivitySvc.List(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
