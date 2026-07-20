package middleware

import (
	"errors"

	"tu-xun/common"
	"tu-xun/controller"

	"github.com/gin-gonic/gin"
)

func CheckRole(min int) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionVal := controller.SessionGet(c, "user-session")
		if sessionVal == nil {
			c.Error(common.ErrNew(errors.New("您未登录"), common.AuthErr))
			c.Abort()
			return
		}
		userSession, ok := sessionVal.(controller.UserSession)
		if !ok || userSession.ID == 0 {
			c.Error(common.ErrNew(errors.New("您未登录"), common.AuthErr))
			c.Abort()
			return
		}
		if userSession.Level < min {
			c.Error(common.ErrNew(errors.New("权限不足"), common.LevelErr))
			c.Abort()
			return
		}
		c.Next()
	}
}
