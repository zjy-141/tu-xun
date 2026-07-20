package controller

import (
	"encoding/gob"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"resp"`
	Message string `json:"message"`
	Code    uint64 `json:"code"`
}

// ResponseNew 构造统一响应结构，自动保存 Session
func ResponseNew(c *gin.Context, obj any) *Response {
	session := sessions.Default(c)
	if session.Save() != nil {
		return &Response{
			Success: false,
			Message: "fail to save session",
			Code:    uint64(common.SysErr),
		}
	}
	return &Response{
		Success: true,
		Data:    obj,
		Message: "",
		Code:    0,
	}
}

var srv = service.New()

func init() {
	gob.Register(UserSession{})
}
