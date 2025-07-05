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
var systemMessageCtl *controller.SystemMessage
var globalSettingsCtl *controller.GlobalSettings
var friendSettingsCtl *controller.FriendSettings
var chatRoomSettingsCtl *controller.ChatRoomSettings
var wechatServerCallbackCtl *controller.WechatServerCallback
var aiCallbackCtl *controller.AICallback
var momentsCtl *controller.Moments
var probeCtl *controller.Probe

func initController() {
	chatHistoryCtl = controller.NewChatHistoryController()
	attachDownloadCtl = controller.NewAttachDownloadController()
	chatRoomCtl = controller.NewChatRoomController()
	contactCtl = controller.NewContactController()
	loginCtl = controller.NewLoginController()
	messageCtl = controller.NewMessageController()
	systemMessageCtl = controller.NewSystemMessageController()
	globalSettingsCtl = controller.NewGlobalSettingsController()
	friendSettingsCtl = controller.NewFriendSettingsController()
	chatRoomSettingsCtl = controller.NewChatRoomSettingsController()
	wechatServerCallbackCtl = controller.NewWechatServerCallbackController()
	aiCallbackCtl = controller.NewAICallbackController()
	momentsCtl = controller.NewMomentsController()
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

	// 登录相关接口
	api.GET("/robot/is-running", loginCtl.IsRunning)
	api.GET("/robot/is-loggedin", loginCtl.IsLoggedIn)
	api.POST("/robot/login", loginCtl.Login)
	api.POST("/robot/login/check", loginCtl.LoginCheck)
	api.DELETE("/robot/logout", loginCtl.Logout)

	// 联系人相关接口
	api.GET("/robot/contacts", contactCtl.GetContacts)
	api.POST("/robot/contact/friend/pass-verify", contactCtl.FriendPassVerify)
	api.POST("/robot/contacts/sync", contactCtl.SyncContact)
	api.DELETE("/robot/contact/friend", contactCtl.FriendDelete)

	// 群聊相关接口
	api.POST("/robot/chat-room/members/sync", chatRoomCtl.SyncChatRoomMember)
	api.POST("/robot/chat-room/join", chatRoomCtl.GroupConsentToJoin)
	api.GET("/robot/chat-room/members", chatRoomCtl.GetChatRoomMembers)
	api.POST("/robot/chat-room/name", chatRoomCtl.GroupSetChatRoomName)
	api.POST("/robot/chat-room/remark", chatRoomCtl.GroupSetChatRoomRemarks)
	api.POST("/robot/chat-room/announcement", chatRoomCtl.GroupSetChatRoomAnnouncement)
	api.DELETE("/robot/chat-room/members", chatRoomCtl.GroupDelChatRoomMember)
	api.DELETE("/robot/chat-room/quit", chatRoomCtl.GroupQuit)

	api.GET("/robot/chat/history", chatHistoryCtl.GetChatHistory)

	api.POST("/robot/message/revoke", messageCtl.MessageRevoke)
	api.POST("/robot/message/send/text", messageCtl.SendTextMessage)
	api.POST("/robot/message/send/image", messageCtl.SendImageMessage)
	api.POST("/robot/message/send/video", messageCtl.SendVideoMessage)
	api.POST("/robot/message/send/voice", messageCtl.SendVoiceMessage)
	api.POST("/robot/message/send/music", messageCtl.SendMusicMessage)

	api.GET("/robot/system-messages", systemMessageCtl.GetRecentMonthMessages)

	api.GET("/robot/chat/image/download", attachDownloadCtl.DownloadImage)
	api.GET("/robot/chat/voice/download", attachDownloadCtl.DownloadVoice)
	api.GET("/robot/chat/file/download", attachDownloadCtl.DownloadFile)
	api.GET("/robot/chat/video/download", attachDownloadCtl.DownloadVideo)

	api.GET("/robot/global-settings", globalSettingsCtl.GetGlobalSettings)
	api.POST("/robot/global-settings", globalSettingsCtl.SaveGlobalSettings)

	api.GET("/robot/friend-settings", friendSettingsCtl.GetFriendSettings)
	api.POST("/robot/friend-settings", friendSettingsCtl.SaveFriendSettings)

	api.GET("/robot/chat-room-settings", chatRoomSettingsCtl.GetChatRoomSettings)
	api.POST("/robot/chat-room-settings", chatRoomSettingsCtl.SaveChatRoomSettings)

	// 朋友圈接口
	api.GET("/robot/moments/list", momentsCtl.FriendCircleGetList)
	api.GET("/robot/moments/down-media", momentsCtl.FriendCircleDownFriendCircleMedia)

	api.POST("/robot/ai-callback/voice/doubao-tts", aiCallbackCtl.DoubaoTTS)

	api.POST("/wechat-client/:wechatID/sync-message", wechatServerCallbackCtl.SyncMessageCallback)
	api.POST("/wechat-client/:wechatID/logout", wechatServerCallbackCtl.LogoutCallback)

	return nil
}
