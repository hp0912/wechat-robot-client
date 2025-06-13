package settings

import (
	"wechat-robot-client/model"

	"gorm.io/datatypes"
)

type AIConfig struct {
	BaseURL         string
	APIKey          string
	Model           string
	Prompt          string
	ImageModel      string
	ImageAISettings datatypes.JSON
}

type Settings interface {
	InitByMessage(message *model.Message) error
	GetAIConfig() AIConfig
	IsAIChatEnabled() bool
	IsAIDrawingEnabled() bool
	IsAITrigger() bool
}
