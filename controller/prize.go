package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Prize struct{}

// MyPrizes 获取我的奖品
func (info *Prize) MyPrizes(c *gin.Context) {
	var params service.MyPrizesParams
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller prize my: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.Prize.MyPrizes(params)
	if err != nil {
		logger.Errorf("controller prize my: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
