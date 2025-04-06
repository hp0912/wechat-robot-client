package main

import (
	"log"
	"wechat-robot-client/startup"
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
}
