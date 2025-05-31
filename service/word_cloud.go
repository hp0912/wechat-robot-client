package service

import (
	"context"
	"log"
	"wechat-robot-client/dto"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
)

type WordCloudService struct {
	ctx context.Context
}

func NewWordCloudService(ctx context.Context) *WordCloudService {
	return &WordCloudService{
		ctx: ctx,
	}
}

func (s *WordCloudService) WordCloudDaily(chatRoomID string, startTime, endTime int64) error {
	msgRespo := repository.NewMessageRepo(s.ctx, vars.DB)
	messages, err := msgRespo.GetMessagesByTimeRange(vars.RobotRuntime.WxID, chatRoomID, startTime, endTime)
	if err != nil {
		return err
	}
	if len(messages) == 0 {
		log.Printf("[词云] 群聊 %s 在昨天没有消息，跳过处理\n", chatRoomID)
		return nil
	}
	resp, err := resty.New().R().
		SetBody(dto.WordCloudRequest{
			ChatRoomID: chatRoomID,
			Content:    "",
			Mode:       "yesterday",
		}).
		Post(vars.WordCloudUrl)
	if err != nil {
		return err
	}

	return nil
}
