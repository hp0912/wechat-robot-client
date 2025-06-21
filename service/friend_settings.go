package service

import (
	"context"
	"strings"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type FriendSettingsService struct {
	ctx            context.Context
	Message        *model.Message
	gsRespo        *repository.GlobalSettings
	fsRespo        *repository.FriendSettings
	globalSettings *model.GlobalSettings
	friendSettings *model.FriendSettings
}

var _ settings.Settings = (*FriendSettingsService)(nil)

func NewFriendSettingsService(ctx context.Context) *FriendSettingsService {
	return &FriendSettingsService{
		ctx:     ctx,
		gsRespo: repository.NewGlobalSettingsRepo(ctx, vars.DB),
		fsRespo: repository.NewFriendSettingsRepo(ctx, vars.DB),
	}
}

func (s *FriendSettingsService) InitByMessage(message *model.Message) error {
	s.Message = message
	globalSettings, err := s.gsRespo.GetGlobalSettings()
	if err != nil {
		return err
	}
	s.globalSettings = globalSettings
	friendSettings, err := s.fsRespo.GetFriendSettings(message.FromWxID)
	if err != nil {
		return err
	}
	s.friendSettings = friendSettings
	return nil
}

func (s *FriendSettingsService) GetAIConfig() settings.AIConfig {
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
		if s.globalSettings.ChatPrompt != "" {
			aiConfig.Prompt = s.globalSettings.ChatPrompt
		}
		if s.globalSettings.ImageModel != "" {
			aiConfig.ImageModel = s.globalSettings.ImageModel
		}
		if s.globalSettings.ImageAISettings != nil {
			aiConfig.ImageAISettings = s.globalSettings.ImageAISettings
		}
	}
	if s.friendSettings != nil {
		if s.friendSettings.ChatBaseURL != nil && *s.friendSettings.ChatBaseURL != "" {
			aiConfig.BaseURL = *s.friendSettings.ChatBaseURL
		}
		if s.friendSettings.ChatAPIKey != nil && *s.friendSettings.ChatAPIKey != "" {
			aiConfig.APIKey = *s.friendSettings.ChatAPIKey
		}
		if s.friendSettings.ChatModel != nil && *s.friendSettings.ChatModel != "" {
			aiConfig.Model = *s.friendSettings.ChatModel
		}
		if s.friendSettings.ChatPrompt != nil && *s.friendSettings.ChatPrompt != "" {
			aiConfig.Prompt = *s.friendSettings.ChatPrompt
		}
		if s.friendSettings.ImageModel != nil && *s.friendSettings.ImageModel != "" {
			aiConfig.ImageModel = *s.friendSettings.ImageModel
		}
		if s.friendSettings.ImageAISettings != nil {
			aiConfig.ImageAISettings = s.friendSettings.ImageAISettings
		}
	}
	aiConfig.BaseURL = strings.TrimRight(aiConfig.BaseURL, "/")
	if !strings.HasSuffix(aiConfig.BaseURL, "/v1") {
		aiConfig.BaseURL += "/v1"
	}
	return aiConfig
}

func (s *FriendSettingsService) IsAIChatEnabled() bool {
	if s.friendSettings != nil && s.friendSettings.ChatAIEnabled != nil {
		return *s.friendSettings.ChatAIEnabled
	}
	if s.globalSettings != nil && s.globalSettings.ChatAIEnabled != nil {
		return *s.globalSettings.ChatAIEnabled
	}
	return false
}

func (s *FriendSettingsService) IsAIDrawingEnabled() bool {
	if s.friendSettings != nil && s.friendSettings.ImageAIEnabled != nil {
		return *s.friendSettings.ImageAIEnabled
	}
	if s.globalSettings != nil && s.globalSettings.ImageAIEnabled != nil {
		return *s.globalSettings.ImageAIEnabled
	}
	return false
}

func (s *FriendSettingsService) IsAITrigger() bool {
	return s.IsAIChatEnabled()
}

func (s *FriendSettingsService) GetAITriggerWord() string {
	return ""
}

func (s *FriendSettingsService) GetFriendSettings(contactID string) (*model.FriendSettings, error) {
	return s.fsRespo.GetFriendSettings(contactID)
}

func (s *FriendSettingsService) SaveFriendSettings(data *model.FriendSettings) error {
	if data.ID == 0 {
		return s.fsRespo.Create(data)
	}
	return s.fsRespo.Update(data)
}
