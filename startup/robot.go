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
	robotRespo := repository.NewRobotAdminRepo(context.Background(), vars.AdminDB)
	robot := robotRespo.GetByRobotID(robotId)
	if robot == nil {
		return errors.New("未找到机器人配置")
	}
	vars.RobotRuntime.RobotID = robot.RobotID
	vars.RobotRuntime.WxID = robot.WxID
	vars.RobotRuntime.DeviceID = robot.DeviceID
	vars.RobotRuntime.DeviceName = robot.DeviceName
	vars.RobotRuntime.ServerHost = robot.ServerHost
	vars.RobotRuntime.ServerPort = robot.ServerPort

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
