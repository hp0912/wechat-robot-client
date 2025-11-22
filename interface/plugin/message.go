package plugin

import (
	"context"
	"io"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"

	"github.com/sashabaranov/go-openai"
)

type MessageServiceIface interface {
	SendTextMessage(toWxID, content string, at ...string) error
	SendLongTextMessage(toWxID string, longText string) error
	SendAppMessage(toWxID string, appMsgType int, appMsgXml string) error
	MsgUploadImg(toWxID string, image io.Reader) (*model.Message, error)
	SendImageMessageByRemoteURL(toWxID string, imageURL string) error
	SendVideoMessageByRemoteURL(toWxID string, videoURL string) error
	MsgSendVoice(toWxID string, voice io.Reader, voiceExt string) error
	MsgSendVideo(toWxID string, video io.Reader, videoExt string) error
	SendMusicMessage(toWxID string, songTitle string) error
	ShareLink(toWxID string, shareLinkInfo robot.ShareLinkMessage) error
	ResetChatRoomAIMessageContext(message *model.Message) error
	GetAIMessageContext(message *model.Message) ([]openai.ChatCompletionMessage, error)
	SetMessageIsInContext(message *model.Message) error
	XmlDecoder(content string) (robot.XmlMessage, error)
	UpdateMessage(message *model.Message) error
	ChatRoomAIDisabled(chatRoomID string) error
}

type MessageContext struct {
	Context        context.Context
	Settings       settings.Settings
	Message        *model.Message
	MessageContent string
	Pat            bool
	ReferMessage   *model.Message
	MessageService MessageServiceIface
}

type MessageHandler interface {
	GetName() string
	GetLabels() []string
	PreAction(ctx *MessageContext) bool
	PostAction(ctx *MessageContext)
	Run(ctx *MessageContext) bool
}
