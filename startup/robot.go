package startup

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

func InitWechatRobot() error {
	// 从机器人管理后台加载机器人配置
	// 这些配置需要先登陆机器人管理后台注册微信机器人才能获得
	robotId := os.Getenv("ROBOT_ID")
	if robotId == "" {
		return errors.New("ROBOT_ID 环境变量未设置")
	}
	id, err := strconv.ParseInt(robotId, 10, 64)
	if err != nil {
		return err
	}
	robotAdmin, err := service.NewAdminService(context.Background()).GetRobotByID(id)
	if err != nil {
		return err
	}
	if robotAdmin == nil {
		return errors.New("未找到机器人配置")
	}
	vars.RobotRuntime.RobotID = robotAdmin.ID
	vars.RobotRuntime.WxID = robotAdmin.WeChatID
	vars.RobotRuntime.DeviceID = robotAdmin.DeviceID
	vars.RobotRuntime.DeviceName = robotAdmin.DeviceName
	vars.RobotRuntime.Status = robotAdmin.Status
	var client *robot.Client
	if vars.WechatServerHost != "" {
		client = robot.NewClient(robot.WechatDomain(vars.WechatServerHost))
	} else {
		client = robot.NewClient(robot.WechatDomain(fmt.Sprintf("server_%s:%d", robotAdmin.RobotCode, 9000)))
	}

	vars.RobotRuntime.Client = client

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
				if vars.RobotRuntime.IsLoggedIn() {
					log.Println("微信机器人已登录")
					vars.RobotRuntime.LoginTime = time.Now().Unix()
					err := service.NewLoginService(context.Background()).Online()
					if err != nil {
						log.Println("微信机器人已登录，启动自动心跳失败:", err)
						return err
					}
				} else {
					log.Println("微信机器人服务端未登录")
					err := service.NewLoginService(context.Background()).Offline()
					if err != nil {
						log.Println("微信机器人服务端未登录，设置离线状态失败:", err)
						return err
					}
				}
				return nil
			} else {
				log.Println("等待微信机器人服务端启动...")
			}
		case <-timeoutTimer.C:
			return errors.New("等待微信机器人服务端启动超时，请检查服务端是否正常运行")
		}
	}
}
