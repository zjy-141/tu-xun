package middleware

import (
	"errors"

	"tu-xun/common"
	"tu-xun/controller"
	"tu-xun/model"

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

		// 检查账号是否被封禁
		var user model.User
		if err := model.DB.Select("status").First(&user, userSession.ID).Error; err == nil {
			if user.Status == "banned" {
				c.Error(common.ErrNew(errors.New("账号已被封禁"), common.AuthErr))
				c.Abort()
				return
			}
		}

		if userSession.Level < min {
			c.Error(common.ErrNew(errors.New("权限不足"), common.LevelErr))
			c.Abort()
			return
		}
		c.Next()
	}
}
