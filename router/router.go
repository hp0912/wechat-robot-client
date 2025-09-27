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
var systemSettingsCtl *controller.SystemSettings
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
	systemSettingsCtl = controller.NewSystemSettingsController()
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
	api.GET("/robot/get-cached-info", loginCtl.GetCachedInfo)
	api.POST("/robot/import-login-data", loginCtl.ImportLoginData)
	api.POST("/robot/login", loginCtl.Login)
	api.POST("/robot/login/check", loginCtl.LoginCheck)
	api.POST("/robot/login/2fa", loginCtl.LoginYPayVerificationcode)
	api.POST("/robot/login/data62", loginCtl.LoginData62Login)
	api.POST("/robot/login/data62-sms-again", loginCtl.LoginData62SMSAgain)
	api.POST("/robot/login/data62-sms-verify", loginCtl.LoginData62SMSVerify)
	api.POST("/robot/login/a16", loginCtl.LoginA16Data)
	api.DELETE("/robot/logout", loginCtl.Logout)

	// 联系人相关接口
	api.GET("/robot/contacts", contactCtl.GetContacts)
	api.POST("/robot/contact/friend/search", contactCtl.FriendSearch)
	api.POST("/robot/contact/friend/add", contactCtl.FriendSendRequest)
	api.POST("/robot/contact/friend/add-from-chat-room", contactCtl.FriendSendRequestFromChatRoom)
	api.POST("/robot/contact/friend/remark", contactCtl.FriendSetRemarks)
	api.POST("/robot/contact/friend/pass-verify", contactCtl.FriendPassVerify)
	api.POST("/robot/contacts/sync", contactCtl.SyncContact)
	api.DELETE("/robot/contact/friend", contactCtl.FriendDelete)

	// 群聊相关接口
	api.POST("/robot/chat-room/members/sync", chatRoomCtl.SyncChatRoomMember)
	api.GET("/robot/chat-room/members", chatRoomCtl.GetChatRoomMembers)
	api.GET("/robot/chat-room/not-left-members", chatRoomCtl.GetNotLeftMembers)
	api.POST("/robot/chat-room/create", chatRoomCtl.CreateChatRoom)
	api.POST("/robot/chat-room/invite", chatRoomCtl.InviteChatRoomMember)
	api.POST("/robot/chat-room/join", chatRoomCtl.GroupConsentToJoin)
	api.POST("/robot/chat-room/name", chatRoomCtl.GroupSetChatRoomName)
	api.POST("/robot/chat-room/remark", chatRoomCtl.GroupSetChatRoomRemarks)
	api.POST("/robot/chat-room/announcement", chatRoomCtl.GroupSetChatRoomAnnouncement)
	api.DELETE("/robot/chat-room/members", chatRoomCtl.GroupDelChatRoomMember)
	api.DELETE("/robot/chat-room/quit", chatRoomCtl.GroupQuit)

	api.GET("/robot/chat/history", chatHistoryCtl.GetChatHistory)

	// 消息相关接口
	api.POST("/robot/message/revoke", messageCtl.MessageRevoke)
	api.POST("/robot/message/send/text", messageCtl.SendTextMessage)
	api.POST("/robot/message/send/image", messageCtl.SendImageMessage)
	api.POST("/robot/message/send/video", messageCtl.SendVideoMessage)
	api.POST("/robot/message/send/voice", messageCtl.SendVoiceMessage)
	api.POST("/robot/message/send/music", messageCtl.SendMusicMessage)
	api.POST("/robot/message/send/file", messageCtl.SendFileMessage)

	// 系统消息相关接口
	api.GET("/robot/system-messages", systemMessageCtl.GetRecentMonthMessages)
	api.POST("/robot/system-messages/mark-as-read", systemMessageCtl.MarkAsReadBatch)

	// 系统设置相关接口
	api.GET("/robot/system-settings", systemSettingsCtl.GetSystemSettings)
	api.POST("/robot/system-settings", systemSettingsCtl.SaveSystemSettings)

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
	api.GET("/robot/moments/sync", momentsCtl.SyncMoments)
	api.GET("/robot/moments/settings", momentsCtl.GetFriendCircleSettings)
	api.GET("/robot/moments/get-detail", momentsCtl.FriendCircleGetDetail)
	api.GET("/robot/moments/get-id-detail", momentsCtl.FriendCircleGetIdDetail)
	api.GET("/robot/moments/down-media", momentsCtl.FriendCircleDownFriendCircleMedia)
	api.POST("/robot/moments/settings", momentsCtl.SaveFriendCircleSettings)
	api.POST("/robot/moments/comment", momentsCtl.FriendCircleComment)
	api.POST("/robot/moments/upload-media", momentsCtl.FriendCircleUpload)
	api.POST("/robot/moments/post", momentsCtl.FriendCirclePost)
	api.POST("/robot/moments/operate", momentsCtl.FriendCircleOperation)
	api.POST("/robot/moments/privacy-settings", momentsCtl.FriendCirclePrivacySettings)

	// 豆包AI回调相关接口
	api.POST("/robot/ai-callback/voice/doubao-tts", aiCallbackCtl.DoubaoTTS)

	// 微信服务端回调相关接口
	api.POST("/wechat-client/:wechatID/sync-message", wechatServerCallbackCtl.SyncMessageCallback)
	api.POST("/wechat-client/:wechatID/logout", wechatServerCallbackCtl.LogoutCallback)

	// 微信小程序相关接口
	api.POST("/robot/wxapp/qrcode-auth-login", controller.NewWXAppController().WxappQrcodeAuthLogin)

	return nil
}
