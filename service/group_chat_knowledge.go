package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
)

// 记录每个群聊最后一次成功提取知识的时间戳，未处理完的消息会在下次继续累积
var groupChatKnowledgeLastTime sync.Map

type GroupChatKnowledgeService struct {
	ctx    context.Context
	msgRepo *repository.Message
	ctRepo  *repository.Contact
	gsRepo  *repository.GlobalSettings
}

func NewGroupChatKnowledgeService(ctx context.Context) *GroupChatKnowledgeService {
	return &GroupChatKnowledgeService{
		ctx:    ctx,
		msgRepo: repository.NewMessageRepo(ctx, vars.DB),
		ctRepo:  repository.NewContactRepo(ctx, vars.DB),
		gsRepo:  repository.NewGlobalSettingsRepo(ctx, vars.DB),
	}
}

type extractedKnowledge struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ExtractGroupChatKnowledge 扫描活跃群聊，对累积文本消息达 50 条的群聊调用 LLM 提取有价值内容并存入知识库
func (s *GroupChatKnowledgeService) ExtractGroupChatKnowledge() error {
	globalSettings, err := s.gsRepo.GetGlobalSettings()
	if err != nil {
		return fmt.Errorf("获取全局设置失败: %w", err)
	}
	if globalSettings == nil || globalSettings.ChatAIEnabled == nil || !*globalSettings.ChatAIEnabled {
		return nil
	}
	if globalSettings.ChatAPIKey == "" || globalSettings.ChatBaseURL == "" {
		return nil
	}

	contacts, err := s.ctRepo.FindRecentChatRoomContacts()
	if err != nil {
		return fmt.Errorf("查询活跃群聊失败: %w", err)
	}

	now := time.Now()

	for _, contact := range contacts {
		chatRoomID := contact.WechatID

		var startTime int64
		if val, ok := groupChatKnowledgeLastTime.Load(chatRoomID); ok {
			startTime = val.(int64)
		} else {
			startTime = now.Add(-1 * time.Hour).Unix()
		}
		endTime := now.Unix()

		messages, err := s.msgRepo.GetChatRoomTextMessagesByTimeRange(chatRoomID, vars.RobotRuntime.WxID, startTime, endTime, 500)
		if err != nil {
			log.Printf("[GroupChatKnowledge] 查询群聊 %s 消息失败: %v", chatRoomID, err)
			continue
		}

		if len(messages) < 50 {
			continue
		}

		chatRoomName := chatRoomID
		if contact.Nickname != nil && *contact.Nickname != "" {
			chatRoomName = *contact.Nickname
		}

		knowledgeItems, err := s.callLLMExtractKnowledge(globalSettings, chatRoomName, messages)
		if err != nil {
			log.Printf("[GroupChatKnowledge] 群聊 %s 提取知识失败: %v", chatRoomName, err)
			continue
		}

		if len(knowledgeItems) == 0 {
			log.Printf("[GroupChatKnowledge] 群聊 %s 无有价值内容", chatRoomName)
			groupChatKnowledgeLastTime.Store(chatRoomID, endTime)
			continue
		}

		for _, item := range knowledgeItems {
			if err := vars.KnowledgeService.AddDocument(s.ctx, item.Title, item.Content, "group_chat:"+chatRoomID, "group_chat"); err != nil {
				log.Printf("[GroupChatKnowledge] 存储知识失败: %v", err)
			}
		}

		log.Printf("[GroupChatKnowledge] 群聊 %s 提取了 %d 条知识", chatRoomName, len(knowledgeItems))
		groupChatKnowledgeLastTime.Store(chatRoomID, endTime)
	}

	return nil
}

func (s *GroupChatKnowledgeService) callLLMExtractKnowledge(globalSettings *model.GlobalSettings, chatRoomName string, messages []*dto.TextMessageItem) ([]extractedKnowledge, error) {
	var content strings.Builder
	for _, msg := range messages {
		timeStr := time.Unix(msg.CreatedAt, 0).Format("15:04:05")
		fmt.Fprintf(&content, "[%s] %s: %s\n", timeStr, msg.Nickname, msg.Message)
	}

	systemPrompt := `你是一个群聊内容分析助手。请分析以下微信群「` + chatRoomName + `」的聊天记录，提取其中有价值的信息。

提取规则：
1. 用户分享的资源（文章、工具、网站、开源项目等）
2. 用户提到的外部链接或推荐
3. 有用的知识点或技术分享
4. 有趣的新闻或行业动态
5. 有价值的观点或经验分享

忽略以下内容：
- 闲聊、寒暄、灌水
- 表情包、重复内容
- 无实质内容的对话

输出 JSON 数组格式，每条知识：
[{"title": "简短标题（20字以内）", "content": "详细内容描述，包含关键信息和上下文"}]

如果没有有价值的信息，返回空数组 []。只返回 JSON，不要其他内容。`

	aiConfig := openai.DefaultConfig(globalSettings.ChatAPIKey)
	aiConfig.BaseURL = utils.NormalizeAIBaseURL(globalSettings.ChatBaseURL)
	client := openai.NewClientWithConfig(aiConfig)

	resp, err := client.CreateChatCompletion(s.ctx, openai.ChatCompletionRequest{
		Model: globalSettings.ChatModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: content.String()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回为空")
	}

	responseContent := strings.TrimSpace(resp.Choices[0].Message.Content)
	responseContent = strings.TrimPrefix(responseContent, "```json")
	responseContent = strings.TrimPrefix(responseContent, "```")
	responseContent = strings.TrimSuffix(responseContent, "```")
	responseContent = strings.TrimSpace(responseContent)

	var items []extractedKnowledge
	if err := json.Unmarshal([]byte(responseContent), &items); err != nil {
		return nil, fmt.Errorf("解析 LLM 返回结果失败: %w", err)
	}

	return items, nil
}
