package startup

import (
	"context"
	"errors"
	"log"
	"os"
	"time"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

func InitWechatRobot() error {
	// 从机器人管理后台加载机器人配置
	// 这些配置需要先登陆机器人管理后台注册微信机器人才能获得
	robotId := os.Getenv("ROBOT_ID")
	if robotId == "" {
		return errors.New("ROBOT_ID 环境变量未设置")
	}
	_ = repository.NewRobotAdminRepo(context.Background(), vars.AdminDB)

	// 检测微信机器人服务端是否启动
	retryInterval := 10 * time.Second
	retryTicker := time.NewTicker(retryInterval)
	defer retryTicker.Stop()

	timeoutTimer := time.NewTimer(vars.RobotStartTimeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-retryTicker.C:
			if vars.RobotRuntime.IsRunning() {
				log.Println("微信机器人服务端已启动")
				return nil
			} else {
				log.Println("等待微信机器人服务端启动...")
			}
		case <-timeoutTimer.C:
			return errors.New("等待微信机器人服务端启动超时，请检查服务端是否正常运行")
		}
	}
}
