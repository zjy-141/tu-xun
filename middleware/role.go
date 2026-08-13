package middleware

import (
	"errors"

	"tu-xun/common"
	"tu-xun/controller"
	"tu-xun/model"

	"github.com/gin-gonic/gin"
)

// XSessionID 从 X-Session-Id 请求头中恢复会话，适配非 Cookie 宿主环境
func XSessionID() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.GetHeader("X-Session-Id")
		if sid == "" {
			c.Next()
			return
		}
		// 如果已有 cookie 会话，跳过
		if controller.SessionGet(c, "user-session") != nil {
			c.Next()
			return
		}
		// 从服务端映射查找用户 id
		userID, ok := controller.GetXSession(sid)
		if !ok {
			c.Next()
			return
		}
		// 根据用户 id 从数据库重建完整会话，保证等级、昵称等字段为最新
		var user model.User
		if err := model.DB.First(&user, userID).Error; err != nil {
			c.Next()
			return
		}
		us := controller.UserSession{
			ID:       user.ID,
			NetID:    user.NetID,
			Username: user.Name,
			Nickname: user.Nickname,
			Level:    user.Level,
		}
		// 注入到当前会话，后续 CheckRole 可正常读取
		controller.SessionSet(c, "user-session", us)
		c.Next()
	}
}

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
