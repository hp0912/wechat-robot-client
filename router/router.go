package router

import (
	"wechat-robot-client/controller"
	"wechat-robot-client/middleware"

	"github.com/gin-gonic/gin"
)

var robotCtl *controller.Robot
var attachDownloadCtl *controller.AttachDownload

func initController() {
	robotCtl = controller.NewRobotController()
	attachDownloadCtl = controller.NewAttachDownloadController()
}

func RegisterRouter(r *gin.Engine) error {
	r.Use(middleware.ErrorRecover)

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
	api.POST("/probe", robotCtl.Probe)

	api.GET("/robot/is-running", robotCtl.IsRunning)
	api.GET("/robot/is-loggedin", robotCtl.IsLoggedIn)
	api.POST("/robot/login", robotCtl.Login)
	api.POST("/robot/login/check", robotCtl.LoginCheck)

	api.GET("/robot/contacts", robotCtl.GetContacts)
	api.POST("/robot/contacts/sync", robotCtl.SyncContact)

	api.GET("/robot/chat-room/members", robotCtl.GetChatRoomMembers)
	api.POST("/robot/chat-room/members/sync", robotCtl.SyncChatRoomMember)

	api.GET("/robot/chat/history", robotCtl.GetChatHistory)
	api.GET("/robot/chat/image/download", attachDownloadCtl.DownloadImage)
	api.GET("/robot/chat/voice/download", attachDownloadCtl.DownloadVoice)
	api.GET("/robot/chat/file/download", attachDownloadCtl.DownloadFile)

	api.DELETE("/robot/logout", robotCtl.Logout)

	return nil
}
