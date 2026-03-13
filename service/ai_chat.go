package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/pkg/mcp"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
)

type AIChatService struct {
	ctx    context.Context
	config settings.Settings
}

func NewAIChatService(ctx context.Context, config settings.Settings) *AIChatService {
	return &AIChatService{
		ctx:    ctx,
		config: config,
	}
}

func (s *AIChatService) Chat(robotCtx mcp.RobotContext, aiMessages []openai.ChatCompletionMessage) (openai.ChatCompletionMessage, error) {
	aiConfig := s.config.GetAIConfig()

	// 提取最后一条用户消息用于 RAG 检索
	var lastUserQuery string
	for i := len(aiMessages) - 1; i >= 0; i-- {
		if aiMessages[i].Role == openai.ChatMessageRoleUser {
			lastUserQuery = aiMessages[i].Content
			if lastUserQuery == "" && len(aiMessages[i].MultiContent) > 0 {
				for _, mc := range aiMessages[i].MultiContent {
					if mc.Type == openai.ChatMessagePartTypeText && mc.Text != "" {
						lastUserQuery = mc.Text
						break
					}
				}
			}
			break
		}
	}

	// 构建系统提示词（含 RAG 增强）
	basePrompt := aiConfig.Prompt
	if basePrompt == "" {
		basePrompt = "你是一个智能助手。"
	}

	// RAG 增强：检索相关记忆、知识库、历史消息
	contactWxID := robotCtx.SenderWxID
	chatRoomID := ""
	if strings.Contains(robotCtx.FromWxID, "@chatroom") {
		chatRoomID = robotCtx.FromWxID
	} else {
		contactWxID = robotCtx.FromWxID
	}

	enhancedPrompt := basePrompt
	if vars.RAGService != nil && lastUserQuery != "" {
		retrieved := vars.RAGService.RetrieveContext(s.ctx, contactWxID, chatRoomID, lastUserQuery)
		enhancedPrompt = vars.RAGService.BuildEnhancedPrompt(basePrompt, retrieved)
	}

	// 组装系统消息
	systemMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: enhancedPrompt,
	}
	if aiConfig.MaxCompletionTokens > 0 {
		systemMessage.Content += fmt.Sprintf("\n\n请注意，每次回答不能超过%d个汉字。", aiConfig.MaxCompletionTokens)
	}
	aiMessages = append([]openai.ChatCompletionMessage{systemMessage}, aiMessages...)

	openaiConfig := openai.DefaultConfig(aiConfig.APIKey)
	openaiConfig.BaseURL = aiConfig.BaseURL
	client := openai.NewClientWithConfig(openaiConfig)
	req := openai.ChatCompletionRequest{
		Model:    aiConfig.Model,
		Messages: aiMessages,
		Stream:   false,
	}
	if aiConfig.MaxCompletionTokens > 0 {
		// req.MaxCompletionTokens = aiConfig.MaxCompletionTokens
	}

	reply, err := vars.MCPService.ChatWithMCPTools(robotCtx, client, req, 0)

	// 异步：记忆提取 + 会话追踪 + 消息向量化
	if err == nil {
		go s.postChatHook(contactWxID, chatRoomID, robotCtx.MessageID, aiMessages, reply)
	}

	return reply, err
}

// postChatHook 在 AI 回复后异步执行记忆提取、会话追踪
func (s *AIChatService) postChatHook(contactWxID, chatRoomID string, msgID int64, aiMessages []openai.ChatCompletionMessage, reply openai.ChatCompletionMessage) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[RAG] postChatHook panic: %v", r)
		}
	}()

	ctx := context.Background()

	// 1. 更新会话追踪
	if vars.MemoryService != nil {
		vars.MemoryService.TouchSession(ctx, contactWxID, chatRoomID, msgID)
	}

	// 2. 从对话中提取记忆（包含本次回复）
	if vars.MemoryService != nil && len(aiMessages) > 0 {
		// 将回复也加入消息列表用于记忆提取
		allMessages := make([]openai.ChatCompletionMessage, 0, len(aiMessages)+1)
		for _, m := range aiMessages {
			if m.Role != openai.ChatMessageRoleSystem {
				allMessages = append(allMessages, m)
			}
		}
		if reply.Content != "" {
			allMessages = append(allMessages, reply)
		}
		vars.MemoryService.ExtractMemoriesFromConversation(contactWxID, chatRoomID, allMessages)
	}
}
