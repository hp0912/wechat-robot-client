package service

import (
	"context"
	"log"
	"strings"
	"wechat-robot-client/dto"
	"wechat-robot-client/repository"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
)

type WordCloudService struct {
	ctx     context.Context
	msgRepo *repository.Message
}

func NewWordCloudService(ctx context.Context) *WordCloudService {
	return &WordCloudService{
		ctx:     ctx,
		msgRepo: repository.NewMessageRepo(ctx, vars.DB),
	}
}

func (s *WordCloudService) WordCloudDaily(chatRoomID, aiTriggerWord string, startTime, endTime int64) ([]byte, error) {
	messages, err := s.msgRepo.GetMessagesByTimeRange(vars.RobotRuntime.WxID, chatRoomID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		log.Printf("[词云] 群聊 %s 在昨天没有消息，跳过处理\n", chatRoomID)
		return nil, nil
	}
	// 使用 strings.Builder 高效拼接字符串
	var builder strings.Builder
	for i, msg := range messages {
		// 去除首尾空格
		content := strings.TrimSpace(msg.Message)
		// 去除艾特，去除AI触发词
		content = utils.TrimAITriggerAll(content, aiTriggerWord)
		// 如果内容为空，跳过
		if content == "" {
			continue
		}
		// 添加内容
		builder.WriteString(content)
		// 如果不是最后一个元素，添加换行符
		if i < len(messages)-1 {
			builder.WriteString("\n")
		}
	}
	resp, err := resty.New().R().
		SetBody(dto.WordCloudRequest{
			ChatRoomID: chatRoomID,
			Content:    builder.String(),
			Mode:       "yesterday",
		}).
		Post(vars.WordCloudUrl)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}
