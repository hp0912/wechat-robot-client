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
	MsgUploadImg(toWxID string, image io.Reader) error
	SendMusicMessage(toWxID string, songTitle string) error
	ResetChatRoomAIMessageContext(message *model.Message) error
	GetAIMessageContext(message *model.Message) ([]openai.ChatCompletionMessage, error)
	SetMessageIsInContext(message *model.Message) error
	XmlDecoder(content string) (robot.XmlMessage, error)
}

type MessageContext struct {
	Context        context.Context
	Settings       settings.Settings
	Message        *model.Message
	ReferMessage   *model.Message
	MessageService MessageServiceIface
}

type MessageHandler func(ctx *MessageContext) bool
