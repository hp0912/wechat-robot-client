package service

import (
	"context"
	"fmt"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type ChatRoomSettingsService struct {
	ctx      context.Context
	gsRespo  *repository.GlobalSettings
	crsRespo *repository.ChatRoomSettings
}

func NewChatRoomSettingsService(ctx context.Context) *ChatRoomSettingsService {
	return &ChatRoomSettingsService{
		ctx:      ctx,
		gsRespo:  repository.NewGlobalSettingsRepo(ctx, vars.DB),
		crsRespo: repository.NewChatRoomSettingsRepo(ctx, vars.DB),
	}
}

func (s *ChatRoomSettingsService) GetChatRoomSettings(chatRoomID string) (*model.ChatRoomSettings, error) {
	return s.crsRespo.GetChatRoomSettings(chatRoomID)
}

func (s *ChatRoomSettingsService) GetChatRoomWelcomeConfig(chatRoomID string) (*model.ChatRoomSettings, error) {
	globalSettings, err := s.gsRespo.GetGlobalSettings()
	if err != nil {
		return nil, err
	}
	if globalSettings == nil {
		return nil, fmt.Errorf("加载全局配置失败")
	}
	chatRoomSetting, err := s.crsRespo.GetChatRoomSettings(chatRoomID)
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

func (s *ChatRoomSettingsService) GetAllEnableChatRank() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	return s.crsRespo.GetAllEnableChatRank()
}

func (s *ChatRoomSettingsService) GetAllEnableAISummary() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	return s.crsRespo.GetAllEnableAISummary()
}

func (s *ChatRoomSettingsService) GetAllEnableGoodMorning() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	return s.crsRespo.GetAllEnableGoodMorning()
}

func (s *ChatRoomSettingsService) GetAllEnableNews() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	return s.crsRespo.GetAllEnableNews()
}

func (s *ChatRoomSettingsService) SaveChatRoomSettings(data *model.ChatRoomSettings) error {
	if data.ID == 0 {
		return s.crsRespo.Create(data)
	}
	return s.crsRespo.Update(data)
}
