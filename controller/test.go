package controller

import (
	"net/http"
	"tu-xun/config"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Test struct {
}

// Login 内部登录测试
func (ctr *Test) Login(c *gin.Context) {
	var params service.TestLoginParams
	if err := c.ShouldBind(&params); err != nil {
		c.Redirect(http.StatusFound, config.Config.OnlineCallback)
		return
	}
	if params.Password != "totoro@tiaozhan" || params.UserID == 0 {
		c.Redirect(http.StatusFound, config.Config.OnlineCallback)
		return
	}
	resp, err := srv.TestSvc.TestLogin(params)
	if err != nil {
		logger.Errorf("service test login: %v\n", err)
		c.Redirect(http.StatusFound, config.Config.OnlineCallback)
		return
	}
	userSession := UserSession{
		ID:       resp.ID,
		NetID:    resp.NetID,
		Username: resp.Username,
		Nickname: resp.Nickname,
		Level:    resp.Level,
	}
	SessionSet(c, "user-session", userSession)

	sessionID := uuid.New().String()
	StoreXSession(sessionID, resp.ID)
	SessionSet(c, "session-id", sessionID)

	c.JSON(http.StatusOK, ResponseNew(c, service.LoginResult{
		UserSummary: resp,
		SessionID:   sessionID,
	}))
}
