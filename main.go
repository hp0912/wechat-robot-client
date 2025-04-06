package main

import (
	"log"
	"os"
	"wechat-robot-client/router"
	"wechat-robot-client/startup"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := startup.LoadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if err := startup.SetupVars(); err != nil {
		log.Fatalf("MySQL连接失败: %v", err)
	}
	// 注册消息处理插件
	startup.RegisterPlugin()
	// 启动HTTP服务
	gin.SetMode(os.Getenv("GIN_MODE"))
	app := gin.Default()
	// 注册路由
	if err := router.RegisterRouter(app); err != nil {
		log.Fatalf("注册路由失败: %v", err)
	}
	// 启动服务
	if err := app.Run(":9000"); err != nil {
		log.Panicf("服务启动失败：%v", err)
	}
}
