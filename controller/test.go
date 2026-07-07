package controller

import (
	"net/http"
	"tu-xun/config"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Test struct {
}

// 内部登录测试
func (t *Test) Login(c *gin.Context) {
	var params service.TsetLoginParams
	if err := c.ShouldBind(&params); err != nil {
		c.Redirect(http.StatusFound, config.Config.OnlineCallback)
		return
	}
	if params.Password != "totoro@tiaozhan" || params.NetID == "" && params.Username == "" {
		c.Redirect(http.StatusFound, config.Config.OnlineCallback)
		return
	}
	resp, err := srv.TestSvc.TestLogin(params)
	if err != nil {
		logger.Errorf("service test login: %v\n", err)
		c.Redirect(http.StatusFound, config.Config.OnlineCallback)
		return
	}
	SessionSet(c, "user-session", UserSession{
		ID:       resp.ID,
		NetID:    resp.NetID,
		Username: resp.Username,
		Nickname: resp.Nickname,
		Level:    resp.Level,
	})
	c.JSON(http.StatusOK, ResponseNew(c, resp))

}
