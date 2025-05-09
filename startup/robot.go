package startup

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
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
	robotRespo := repository.NewRobotAdminRepo(context.Background(), vars.AdminDB)
	robotAdmin := robotRespo.GetByRobotID(id)
	if robotAdmin == nil {
		return errors.New("未找到机器人配置")
	}
	vars.RobotRuntime.RobotID = robotAdmin.ID
	vars.RobotRuntime.WxID = robotAdmin.WeChatID
	vars.RobotRuntime.DeviceID = robotAdmin.DeviceID
	vars.RobotRuntime.DeviceName = robotAdmin.DeviceName
	vars.RobotRuntime.Status = robotAdmin.Status
	client := robot.NewClient(robot.WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003))) // TODO
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
				if vars.RobotRuntime.Status == model.RobotStatusOnline {
					go service.NewLoginService(context.Background()).HeartbeatStart()
					log.Println("微信机器人已经登陆，开始心跳检测...")
					go service.NewMessageService(context.Background()).SyncMessageStart()
					log.Println("开始同步消息...")
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
