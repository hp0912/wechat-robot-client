package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"wechat-robot-client/common_cron"
	"wechat-robot-client/pkg/shutdown"
	"wechat-robot-client/router"
	"wechat-robot-client/startup"
	"wechat-robot-client/vars"

	"github.com/gin-gonic/gin"
)

var Version = "unknown"

func main() {
	log.Printf("[微信机器人]启动 版本: %s", Version)

	// 加载配置
	if err := startup.LoadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if err := startup.SetupVars(); err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	shutdownManager := shutdown.NewShutdownManager(30 * time.Second)
	// 注册消息处理插件
	startup.RegisterMessagePlugin()
	// 初始化微信机器人
	if err := startup.InitWechatRobot(); err != nil {
		log.Fatalf("启动微信机器人失败: %v", err)
	}
	// 初始化定时任务
	vars.CronManager = common_cron.NewCronManager()
	vars.CronManager.Clear()
	vars.CronManager.Start()
	// 启动HTTP服务
	gin.SetMode(os.Getenv("GIN_MODE"))
	app := gin.Default()

	// 注册路由
	if err := router.RegisterRouter(app); err != nil {
		log.Fatalf("注册路由失败: %v", err)
	}
	dbConn := &shutdown.DBConnection{
		DB:      vars.DB,
		AdminDB: vars.AdminDB,
	}
	redisConn := &shutdown.RedisConnection{
		Client: vars.RedisClient,
	}
	shutdownManager.Register(dbConn)
	shutdownManager.Register(redisConn)
	shutdownManager.Register(vars.RobotRuntime)
	shutdownManager.Register(vars.CronManager)
	// 开始监听停止信号
	shutdownManager.Start()
	// 启动服务
	if err := app.Run(fmt.Sprintf(":%s", vars.WechatClientPort)); err != nil {
		log.Panicf("服务启动失败：%v", err)
	}
}
