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
	ImageModel      model.ImageModel
	ImageAISettings datatypes.JSON
	TTSSettings     datatypes.JSON
	LTTSSettings    datatypes.JSON
}

type Settings interface {
	InitByMessage(message *model.Message) error
	GetAIConfig() AIConfig
	IsAIChatEnabled() bool
	IsAIDrawingEnabled() bool
	IsTTSEnabled() bool
	IsAITrigger() bool
	GetAITriggerWord() string
}
