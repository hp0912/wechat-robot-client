package settings

import (
	"wechat-robot-client/model"

	"gorm.io/datatypes"
)

type AIConfig struct {
	BaseURL               string
	APIKey                string
	Model                 string
	WorkflowModel         string
	ImageRecognitionModel string
	Prompt                string
	MaxCompletionTokens   int
	ImageAISettings       datatypes.JSON
	TTSSettings           datatypes.JSON
	LTTSSettings          datatypes.JSON
}

type PatConfig struct {
	PatEnabled     bool
	PatType        model.PatType
	PatText        string
	PatVoiceTimbre string
}

type Settings interface {
	InitByMessage(message *model.Message) error
	GetAIConfig() AIConfig
	IsAIChatEnabled() bool
	IsAIDrawingEnabled() bool
	IsTTSEnabled() bool
	IsShortVideoParsingEnabled() bool
	IsAITrigger() bool
	GetAITriggerWord() string
	GetPatConfig() PatConfig
}
