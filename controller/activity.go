package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/logger"

	"github.com/gin-gonic/gin"
)

type Activity struct{}

// CurrentActivity 获取当前活动
func (a *Activity) CurrentActivity(c *gin.Context) {

	resp, err := srv.ActivitySvc.Current()
	if err != nil {
		logger.Errorf("service activity current: %v\n", err)
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// HistoryActivity 获取往期活动列表（分页）
func (a *Activity) HistoryActivity(c *gin.Context) {
	var params common.PagerForm
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller activity history: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.ActivitySvc.History(params)
	if err != nil {
		logger.Errorf("service activity history: %v\n", err)
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
