package service

import (
	"context"
	"strings"
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

func NewFriendSettingsService(ctx context.Context) *FriendSettingsService {
	return &FriendSettingsService{
		ctx:     ctx,
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

func (s *FriendSettingsService) GetAIConfig() AIConfig {
	aiConfig := AIConfig{}
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
	}
	aiConfig.BaseURL = strings.TrimRight(aiConfig.BaseURL, "/")
	if !strings.HasSuffix(aiConfig.BaseURL, "/v1") {
		aiConfig.BaseURL += "/v1"
	}
	return aiConfig
}

func (s *FriendSettingsService) IsAIEnabled() bool {
	if s.friendSettings != nil && s.friendSettings.ChatAIEnabled != nil {
		return *s.friendSettings.ChatAIEnabled
	}
	if s.globalSettings != nil && s.globalSettings.ChatAIEnabled != nil {
		return *s.globalSettings.ChatAIEnabled
	}
	return false
}

func (s *FriendSettingsService) IsAITrigger() bool {
	return s.IsAIEnabled()
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
