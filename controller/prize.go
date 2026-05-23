package controller

import (
	"net/http"
	"tu-xun/logger"

	"github.com/gin-gonic/gin"
)

type Prize struct{}

// MyPrizes 获取我的奖品
func (info *Prize) MyPrizes(c *gin.Context) {
	resp, err := srv.Prize.MyPrizes(SessionGet(c, "user-session").(UserSession).ID)
	if err != nil {
		logger.Errorf("controller prize my: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
