package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/repository"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"
)

type ChatRoomSettingsService struct {
	ctx              context.Context
	Message          *model.Message
	gsRepo           *repository.GlobalSettings
	crsRepo          *repository.ChatRoomSettings
	globalSettings   *model.GlobalSettings
	chatRoomSettings *model.ChatRoomSettings
}

var _ settings.Settings = (*ChatRoomSettingsService)(nil)

func NewChatRoomSettingsService(ctx context.Context) *ChatRoomSettingsService {
	return &ChatRoomSettingsService{
		ctx:     ctx,
		gsRepo:  repository.NewGlobalSettingsRepo(ctx, vars.DB),
		crsRepo: repository.NewChatRoomSettingsRepo(ctx, vars.DB),
	}
}

func (s *ChatRoomSettingsService) GetChatRoomSettings(chatRoomID string) (*model.ChatRoomSettings, error) {
	return s.crsRepo.GetChatRoomSettings(chatRoomID)
}

func (s *ChatRoomSettingsService) InitByMessage(message *model.Message) error {
	s.Message = message
	globalSettings, err := s.gsRepo.GetGlobalSettings()
	if err != nil {
		return err
	}
	s.globalSettings = globalSettings
	chatRoomSettings, err := s.crsRepo.GetChatRoomSettings(message.FromWxID)
	if err != nil {
		return err
	}
	s.chatRoomSettings = chatRoomSettings
	return nil
}

func (s *ChatRoomSettingsService) GetAIConfig() settings.AIConfig {
	aiConfig := settings.AIConfig{}
	if s.globalSettings != nil {
		if s.globalSettings.ChatBaseURL != "" {
			aiConfig.BaseURL = s.globalSettings.ChatBaseURL
		}
		if s.globalSettings.ChatAPIKey != "" {
			aiConfig.APIKey = s.globalSettings.ChatAPIKey
		}
		if s.globalSettings.ChatModel != "" {
			aiConfig.Model = s.globalSettings.ChatModel
		}
		if s.globalSettings.ImageRecognitionModel != "" {
			aiConfig.ImageRecognitionModel = s.globalSettings.ImageRecognitionModel
		}
		if s.globalSettings.ChatPrompt != "" {
			aiConfig.Prompt = s.globalSettings.ChatPrompt
		}
		if s.globalSettings.MaxCompletionTokens != nil {
			aiConfig.MaxCompletionTokens = *s.globalSettings.MaxCompletionTokens
		}
		if s.globalSettings.ImageModel != "" {
			aiConfig.ImageModel = s.globalSettings.ImageModel
		}
		if s.globalSettings.ImageAISettings != nil {
			aiConfig.ImageAISettings = s.globalSettings.ImageAISettings
		}
		if s.globalSettings.TTSSettings != nil {
			aiConfig.TTSSettings = s.globalSettings.TTSSettings
		}
		if s.globalSettings.LTTSSettings != nil {
			aiConfig.LTTSSettings = s.globalSettings.LTTSSettings
		}
	}
	if s.chatRoomSettings != nil {
		if s.chatRoomSettings.ChatBaseURL != nil && *s.chatRoomSettings.ChatBaseURL != "" {
			aiConfig.BaseURL = *s.chatRoomSettings.ChatBaseURL
		}
		if s.chatRoomSettings.ChatAPIKey != nil && *s.chatRoomSettings.ChatAPIKey != "" {
			aiConfig.APIKey = *s.chatRoomSettings.ChatAPIKey
		}
		if s.chatRoomSettings.ChatModel != nil && *s.chatRoomSettings.ChatModel != "" {
			aiConfig.Model = *s.chatRoomSettings.ChatModel
		}
		if s.chatRoomSettings.ImageRecognitionModel != nil && *s.chatRoomSettings.ImageRecognitionModel != "" {
			aiConfig.ImageRecognitionModel = *s.chatRoomSettings.ImageRecognitionModel
		}
		if s.chatRoomSettings.ChatPrompt != nil && *s.chatRoomSettings.ChatPrompt != "" {
			aiConfig.Prompt = *s.chatRoomSettings.ChatPrompt
		}
		if s.chatRoomSettings.MaxCompletionTokens != nil {
			aiConfig.MaxCompletionTokens = *s.chatRoomSettings.MaxCompletionTokens
		}
		if s.chatRoomSettings.ImageModel != nil && *s.chatRoomSettings.ImageModel != "" {
			aiConfig.ImageModel = *s.chatRoomSettings.ImageModel
		}
		if s.chatRoomSettings.ImageAISettings != nil {
			aiConfig.ImageAISettings = s.chatRoomSettings.ImageAISettings
		}
		if s.chatRoomSettings.TTSSettings != nil {
			aiConfig.TTSSettings = s.chatRoomSettings.TTSSettings
		}
		if s.chatRoomSettings.LTTSSettings != nil {
			aiConfig.LTTSSettings = s.chatRoomSettings.LTTSSettings
		}
	}
	aiConfig.BaseURL = utils.NormalizeAIBaseURL(aiConfig.BaseURL)
	return aiConfig
}

func (s *ChatRoomSettingsService) IsAIChatEnabled() bool {
	if s.chatRoomSettings != nil && s.chatRoomSettings.ChatAIEnabled != nil {
		return *s.chatRoomSettings.ChatAIEnabled
	}
	if s.globalSettings != nil && s.globalSettings.ChatAIEnabled != nil {
		return *s.globalSettings.ChatAIEnabled
	}
	return false
}

func (s *ChatRoomSettingsService) IsAIDrawingEnabled() bool {
	if s.chatRoomSettings != nil && s.chatRoomSettings.ImageAIEnabled != nil {
		return *s.chatRoomSettings.ImageAIEnabled
	}
	if s.globalSettings != nil && s.globalSettings.ImageAIEnabled != nil {
		return *s.globalSettings.ImageAIEnabled
	}
	return false
}

func (s *ChatRoomSettingsService) IsTTSEnabled() bool {
	if s.chatRoomSettings != nil && s.chatRoomSettings.TTSEnabled != nil {
		return *s.chatRoomSettings.TTSEnabled
	}
	if s.globalSettings != nil && s.globalSettings.TTSEnabled != nil {
		return *s.globalSettings.TTSEnabled
	}
	return false
}

func (s *ChatRoomSettingsService) IsShortVideoParsingEnabled() bool {
	if s.chatRoomSettings != nil && s.chatRoomSettings.ShortVideoParsingEnabled != nil {
		return *s.chatRoomSettings.ShortVideoParsingEnabled
	}
	return false
}

func (s *ChatRoomSettingsService) IsAITrigger() bool {
	messageContent := s.Message.Content
	if s.Message.AppMsgType == model.AppMsgTypequote {
		var xmlMessage robot.XmlMessage
		if err := vars.RobotRuntime.XmlDecoder(messageContent, &xmlMessage); err == nil {
			messageContent = xmlMessage.AppMsg.Title
		}
	}
	if s.Message.IsAtMe {
		// 是否是 @所有人
		atAllRegex := regexp.MustCompile(vars.AtAllRegexp)
		if atAllRegex.MatchString(messageContent) {
			// 如果是 @所有人，则不处理
			return false
		}
		return true
	}
	if s.chatRoomSettings == nil {
		if s.globalSettings == nil {
			return false
		}
		if s.globalSettings.ChatAIEnabled == nil || !*s.globalSettings.ChatAIEnabled {
			return false
		}
		return *s.globalSettings.ChatAITrigger != "" && strings.HasPrefix(messageContent, *s.globalSettings.ChatAITrigger)
	}
	if s.chatRoomSettings.ChatAIEnabled == nil || !*s.chatRoomSettings.ChatAIEnabled {
		return false
	}
	if s.chatRoomSettings.ChatAITrigger != nil && *s.chatRoomSettings.ChatAITrigger != "" {
		return *s.chatRoomSettings.ChatAITrigger != "" && strings.HasPrefix(messageContent, *s.chatRoomSettings.ChatAITrigger)
	}
	return s.globalSettings != nil && s.globalSettings.ChatAITrigger != nil && *s.globalSettings.ChatAITrigger != "" &&
		strings.HasPrefix(messageContent, *s.globalSettings.ChatAITrigger)
}

func (s *ChatRoomSettingsService) GetAITriggerWord() string {
	if s.chatRoomSettings != nil && s.chatRoomSettings.ChatAITrigger != nil && *s.chatRoomSettings.ChatAITrigger != "" {
		return *s.chatRoomSettings.ChatAITrigger
	}
	if s.globalSettings != nil && s.globalSettings.ChatAITrigger != nil && *s.globalSettings.ChatAITrigger != "" {
		return *s.globalSettings.ChatAITrigger
	}
	return ""
}

func (s *ChatRoomSettingsService) GetChatRoomWelcomeConfig(chatRoomID string) (*model.ChatRoomSettings, error) {
	globalSettings, err := s.gsRepo.GetGlobalSettings()
	if err != nil {
		return nil, err
	}
	if globalSettings == nil {
		return nil, fmt.Errorf("加载全局配置失败")
	}
	chatRoomSetting, err := s.crsRepo.GetChatRoomSettings(chatRoomID)
	if err != nil {
		return nil, err
	}
	if chatRoomSetting == nil {
		return &model.ChatRoomSettings{
			WelcomeEnabled:  globalSettings.WelcomeEnabled,
			WelcomeType:     globalSettings.WelcomeType,
			WelcomeText:     globalSettings.WelcomeText,
			WelcomeEmojiMD5: globalSettings.WelcomeEmojiMD5,
			WelcomeEmojiLen: globalSettings.WelcomeEmojiLen,
			WelcomeImageURL: globalSettings.WelcomeImageURL,
			WelcomeURL:      globalSettings.WelcomeURL,
		}, nil
	}
	return chatRoomSetting, nil
}

func (s *ChatRoomSettingsService) GetPatConfig() settings.PatConfig {
	if s.chatRoomSettings != nil {
		if s.chatRoomSettings.PatEnabled != nil {
			return settings.PatConfig{
				PatEnabled:     *s.chatRoomSettings.PatEnabled,
				PatType:        s.chatRoomSettings.PatType,
				PatText:        s.chatRoomSettings.PatText,
				PatVoiceTimbre: s.chatRoomSettings.PatVoiceTimbre,
			}
		}
	}
	if s.globalSettings != nil {
		if s.globalSettings.PatEnabled != nil {
			return settings.PatConfig{
				PatEnabled:     *s.globalSettings.PatEnabled,
				PatType:        s.globalSettings.PatType,
				PatText:        s.globalSettings.PatText,
				PatVoiceTimbre: s.globalSettings.PatVoiceTimbre,
			}
		}
	}
	return settings.PatConfig{}
}

func (s *ChatRoomSettingsService) GetLeaveChatRoomConfig(chatRoomID string) *model.ChatRoomSettings {
	globalSettings, err := s.gsRepo.GetGlobalSettings()
	if err != nil {
		return nil
	}
	chatRoomSettings, err := s.crsRepo.GetChatRoomSettings(chatRoomID)
	if err != nil {
		return nil
	}
	if chatRoomSettings != nil {
		return chatRoomSettings
	}
	if globalSettings != nil {
		return &model.ChatRoomSettings{
			LeaveChatRoomAlertEnabled: globalSettings.LeaveChatRoomAlertEnabled,
			LeaveChatRoomAlertText:    globalSettings.LeaveChatRoomAlertText,
		}
	}
	return nil
}

func (s *ChatRoomSettingsService) GetAllEnableChatRank() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	return s.crsRepo.GetAllEnableChatRank()
}

func (s *ChatRoomSettingsService) GetAllEnableAISummary() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	return s.crsRepo.GetAllEnableAISummary()
}

func (s *ChatRoomSettingsService) GetAllEnableGoodMorning() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	return s.crsRepo.GetAllEnableGoodMorning()
}

func (s *ChatRoomSettingsService) GetAllEnableNews() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	return s.crsRepo.GetAllEnableNews()
}

func (s *ChatRoomSettingsService) SaveChatRoomSettings(data *model.ChatRoomSettings) error {
	if data.ID == 0 {
		return s.crsRepo.Create(data)
	}
	return s.crsRepo.Update(data)
}
