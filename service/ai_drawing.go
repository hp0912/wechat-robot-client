package service

import (
	"context"
	"fmt"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"
)

type AIDrawingService struct {
	ctx    context.Context
	config settings.Settings
}

func NewAIDrawingService(ctx context.Context, config settings.Settings) *AIDrawingService {
	return &AIDrawingService{
		ctx:    ctx,
		config: config,
	}
}

func (s *AIDrawingService) SetAISession(message *model.Message) error {
	return vars.RedisClient.Set(s.ctx, s.GetSessionID(message), true, defaultTTL).Err()
}

func (s *AIDrawingService) RenewAISession(message *model.Message) error {
	return vars.RedisClient.Expire(s.ctx, s.GetSessionID(message), defaultTTL).Err()
}

func (s *AIDrawingService) ExpireAISession(message *model.Message) error {
	return vars.RedisClient.Del(s.ctx, s.GetSessionID(message)).Err()
}

func (s *AIDrawingService) ExpireAllAISessionByChatRoomID(chatRoomID string) error {
	sessionID := fmt.Sprintf("ai_drawing_session_%s:", chatRoomID)
	keys, err := vars.RedisClient.Keys(s.ctx, sessionID+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return vars.RedisClient.Del(s.ctx, keys...).Err()
}

func (s *AIDrawingService) IsInAISession(message *model.Message) (bool, error) {
	cnt, err := vars.RedisClient.Exists(s.ctx, s.GetSessionID(message)).Result()
	return cnt == 1, err
}

func (s *AIDrawingService) GetSessionID(message *model.Message) string {
	return fmt.Sprintf("ai_drawing_session_%s:%s", message.FromWxID, message.SenderWxID)
}

func (s *AIDrawingService) IsAISessionStart(message *model.Message) bool {
	if message.Content == "#进入AI绘图" {
		err := s.SetAISession(message)
		return err == nil
	}
	return false
}

func (s *AIDrawingService) GetAISessionStartTips() string {
	return "AI绘图已开始，请输入您的绘图提示词。10分钟不说话会话将自动结束，您也可以输入 #退出AI绘图 来结束会话。"
}

func (s *AIDrawingService) IsAISessionEnd(message *model.Message) bool {
	if message.Content == "#退出AI绘图" {
		err := s.ExpireAISession(message)
		return err == nil
	}
	return false
}

func (s *AIDrawingService) GetAISessionEndTips() string {
	return "AI绘图已结束，您可以输入 #进入AI绘图 来重新开始。"
}

var _ ai.AIService = (*AIDrawingService)(nil)
