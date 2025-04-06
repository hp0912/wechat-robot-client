package router

import (
	"wechat-robot-client/controller"

	"github.com/gin-gonic/gin"
)

var robotCtl *controller.Robot

func initController() {
	robotCtl = controller.NewRobotController()
}

func RegisterRouter(r *gin.Engine) error {
	initController()

	// 设置信任的内网 IP 范围
	err := r.SetTrustedProxies([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	})
	if err != nil {
		return err
	}

	api := r.Group("/api/v1")
	api.POST("/test", robotCtl.Test)

	return nil
}
