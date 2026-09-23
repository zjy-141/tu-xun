package controller

import (
	"net/http"
	"tu-xun/logger"

	"github.com/gin-gonic/gin"
)

type Box struct{}

// AutoApprove 自动审批状态查询
func (b *Box) AutoApprove(c *gin.Context) {

	resp, err := srv.BoxSvc.AutoApprove()
	if err != nil {
		logger.Errorf("service box auto approve: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
