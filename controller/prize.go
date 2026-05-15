package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Prize struct{}

// MyPrizes 获取我的奖品
func (p *Prize) MyPrizes(c *gin.Context) {
	data, err := srv.Prize.MyPrizes(SessionGet(c, "user-session").(UserSession).ID)
	if err != nil {
		fmt.Printf("controller prize my: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}
