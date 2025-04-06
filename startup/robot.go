package startup

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
)

func IsRunning() bool {
	client := resty.New()
	resp, err := client.R().Get(fmt.Sprintf("%s/IsRunning", vars.RobotRuntime.Doman()))
	if err != nil || resp.StatusCode() != http.StatusOK {
		log.Printf("Error checking if robot is running: %v, http code: %d", err, resp.StatusCode())
		return false
	}
	return resp.String() == "OK"
}

func InitWechatRobot() error {
	// 从机器人管理后台加载机器人配置
	// 这些配置需要先登陆机器人管理后台注册微信机器人才能获得
	// TODO

	// 检测微信机器人服务端是否启动
	retryInterval := 10 * time.Second
	retryTicker := time.NewTicker(retryInterval)
	defer retryTicker.Stop()

	timeoutTimer := time.NewTimer(vars.RobotStartTimeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-retryTicker.C:
			if IsRunning() {
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
