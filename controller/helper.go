package controller

import (
	"fmt"
	"tu-xun/common"

	"github.com/gin-gonic/gin"
)

// getCurrentUser 从 session 获取当前登录用户，未登录时自动写入错误并返回 nil
func getCurrentUser(c *gin.Context) *UserSession {
	session := SessionGet(c, "user-session")
	if session == nil {
		c.Error(common.ErrNew(fmt.Errorf("您未登录"), common.AuthErr))
		c.Abort()
		return nil
	}
	us, ok := session.(UserSession)
	if !ok {
		c.Error(common.ErrNew(fmt.Errorf("会话数据异常"), common.AuthErr))
		c.Abort()
		return nil
	}
	return &us
}
