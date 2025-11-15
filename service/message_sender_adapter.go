package service

import "wechat-robot-client/pkg/mcp"

type MessageSenderAdapter struct {
	messageService *MessageService
}

func NewMessageSenderAdapter(messageService *MessageService) mcp.MessageSender {
	return &MessageSenderAdapter{
		messageService: messageService,
	}
}

func (a *MessageSenderAdapter) SendTextMessage(toWxID, content string, at ...string) error {
	return a.messageService.SendTextMessage(toWxID, content, at...)
}

func (a *MessageSenderAdapter) SendAppMessage(toWxID string, appMsgType int, appMsgXml string) error {
	return a.messageService.SendAppMessage(toWxID, appMsgType, appMsgXml)
}
