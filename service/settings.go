package service

import "wechat-robot-client/model"

type AIConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Prompt  string
}

type Settings interface {
	InitByMessage(message *model.Message) error
	GetAIConfig() AIConfig
	IsAIEnabled() bool
	IsAITrigger() bool
}
