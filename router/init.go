package router

import (
	"net/http"
	"tu-xun/config"
	"tu-xun/controller"
	"tu-xun/middleware"

	"github.com/gin-gonic/gin"
)

func NewServer() *http.Server {
	// 用 gin.New() 而非 gin.Default()：Default 会自带 gin.Logger()+gin.Recovery()，
	// 而 InitRouter 里已注册 middleware.GinLogger()/GinRecovery(true)（接入 logrus，
	// 含请求上下文与 panic 堆栈），两套并存会导致访问日志重复、Recovery 互相覆盖。
	r := gin.New()
	config.SetCORS(r)
	config.InitSession(r)
	r.Use(middleware.XSessionID())
	r.Static("/uploads", "./uploads")
	InitRouter(r)
	s := &http.Server{
		Addr:    "0.0.0.0:8088",
		Handler: r,
	}
	return s

}

var ctr = controller.New()
