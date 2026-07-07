package main

import (
	"tu-xun/config"
	"tu-xun/logger"
	"tu-xun/router"

	"github.com/gin-gonic/gin"
)

// main 应用入口，初始化 Gin 模式并启动 HTTP 服务
func main() {
	gin.SetMode(config.Config.AppMode)
	srv := router.NewServer()

	if err := srv.ListenAndServe(); err != nil {
		logger.Errorf("fail to init server: %s\n", err.Error())
		panic(err)
	}
}
