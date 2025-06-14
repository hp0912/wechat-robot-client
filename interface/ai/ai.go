package ai

import "wechat-robot-client/model"

type AIService interface {
	GetSessionID(message *model.Message) string
	SetAISession(message *model.Message) error
	RenewAISession(message *model.Message) error
	ExpireAISession(message *model.Message) error
	ExpireAllAISessionByChatRoomID(chatRoomID string) error
	IsInAISession(message *model.Message) (bool, error)
	IsAISessionStart(message *model.Message) bool
	GetAISessionStartTips() string
	IsAISessionEnd(message *model.Message) bool
	GetAISessionEndTips() string
}
