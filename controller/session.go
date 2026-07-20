package controller

import (
	"encoding/gob"

	"tu-xun/logger"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type UserSession struct {
	ID       int64
	NetID    string
	Username string
	Nickname string
	Level    int
}

// _SessionSave 持久化保存 Session，失败时记录错误日志
func _SessionSave(ss sessions.Session) {
	if err := ss.Save(); err != nil {
		logger.Errorf("session save error: %v", err)
	}
}

// SessionGet 从当前请求上下文中获取 Session 值
func SessionGet(c *gin.Context, name string) any {
	session := sessions.Default(c)
	return session.Get(name)
}

// SessionSet 写入并持久化 Session 值
func SessionSet(c *gin.Context, name string, body any) {
	session := sessions.Default(c)
	if body == nil {
		return
	}
	gob.Register(body)
	session.Set(name, body)
	_SessionSave(session)
}

// SessionUpdate 更新 Session 值（等同 SessionSet）
func SessionUpdate(c *gin.Context, name string, body any) {
	SessionSet(c, name, body)
}

// SessionClear 清空当前 Session 所有数据
func SessionClear(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	_SessionSave(session)
}

// SessionDelete 删除指定名称的 Session 键
func SessionDelete(c *gin.Context, name string) {
	session := sessions.Default(c)
	session.Delete(name)
	_SessionSave(session)
}
