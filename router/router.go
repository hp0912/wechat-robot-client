package router

import (
	"wechat-robot-client/controller"
	"wechat-robot-client/middleware"

	"github.com/gin-gonic/gin"
)

var chatHistoryCtl *controller.ChatHistory
var attachDownloadCtl *controller.AttachDownload
var chatRoomCtl *controller.ChatRoom
var contactCtl *controller.Contact
var loginCtl *controller.Login
var messageCtl *controller.Message
var globalSettingsCtl *controller.GlobalSettings
var probeCtl *controller.Probe

func initController() {
	chatHistoryCtl = controller.NewChatHistoryController()
	attachDownloadCtl = controller.NewAttachDownloadController()
	chatRoomCtl = controller.NewChatRoomController()
	contactCtl = controller.NewContactController()
	loginCtl = controller.NewLoginController()
	messageCtl = controller.NewMessageController()
	globalSettingsCtl = controller.NewGlobalSettingsController()
	probeCtl = controller.NewProbeController()
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
	api.POST("/probe", probeCtl.Probe)

	api.GET("/robot/is-running", loginCtl.IsRunning)
	api.GET("/robot/is-loggedin", loginCtl.IsLoggedIn)
	api.POST("/robot/login", loginCtl.Login)
	api.POST("/robot/login/check", loginCtl.LoginCheck)
	api.DELETE("/robot/logout", loginCtl.Logout)

	api.GET("/robot/contacts", contactCtl.GetContacts)
	api.POST("/robot/contacts/sync", contactCtl.SyncContact)

	api.GET("/robot/chat-room/members", chatRoomCtl.GetChatRoomMembers)
	api.POST("/robot/chat-room/members/sync", chatRoomCtl.SyncChatRoomMember)

	api.GET("/robot/chat/history", chatHistoryCtl.GetChatHistory)

	api.POST("/robot/message/revoke", messageCtl.MessageRevoke)
	api.POST("/robot/message/send/text", messageCtl.SendTextMessage)
	api.POST("/robot/message/send/image", messageCtl.SendImageMessage)
	api.POST("/robot/message/send/video", messageCtl.SendVideoMessage)
	api.POST("/robot/message/send/voice", messageCtl.SendVoiceMessage)
	api.POST("/robot/message/send/music", messageCtl.SendMusicMessage)

	api.GET("/robot/chat/image/download", attachDownloadCtl.DownloadImage)
	api.GET("/robot/chat/voice/download", attachDownloadCtl.DownloadVoice)
	api.GET("/robot/chat/file/download", attachDownloadCtl.DownloadFile)
	api.GET("/robot/chat/video/download", attachDownloadCtl.DownloadVideo)

	api.GET("/robot/global-settings", globalSettingsCtl.GetGlobalSettings)
	api.POST("/robot/global-settings", globalSettingsCtl.SaveGlobalSettings)

	return nil
}
