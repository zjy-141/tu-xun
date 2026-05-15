package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Prize struct{}

// MyPrizes 获取我的奖品
func (p *Prize) MyPrizes(c *gin.Context) {
	us := getCurrentUser(c)
	if us == nil {
		return
	}

	data, err := srv.Prize.MyPrizes(int64(us.ID))
	if err != nil {
		fmt.Printf("controller prize my: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}
