package service

import (
	"context"
	"log"
	"regexp"
	"strings"
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

func (s *WordCloudService) WordCloudDaily(chatRoomID string, startTime, endTime int64) ([]byte, error) {
	msgRespo := repository.NewMessageRepo(s.ctx, vars.DB)
	messages, err := msgRespo.GetMessagesByTimeRange(chatRoomID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		log.Printf("[词云] 群聊 %s 在昨天没有消息，跳过处理\n", chatRoomID)
		return nil, nil
	}
	// 正则表达式编译一次，提高性能
	re := regexp.MustCompile(`@([^ | ]+)`)
	// 使用 strings.Builder 高效拼接字符串
	var builder strings.Builder
	for i, msg := range messages {
		// 去除首尾空格
		content := strings.TrimSpace(msg.Message)
		// 去除特殊字符
		content = re.ReplaceAllString(content, "")
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
