package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/pkg/mcp"
	"wechat-robot-client/repository"
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
	systemMessage.Content += "\n\n如果外部工具返回以下结构化标签，你必须原样逐字返回，不能总结、解释、改写、翻译、补充代码块，也不能省略、合并或调整顺序：\n<wechat-robot-text>...</wechat-robot-text>\n<wechat-robot-image-url>...</wechat-robot-image-url>\n<wechat-robot-video-url>...</wechat-robot-video-url>\n<wechat-robot-voice-url>...</wechat-robot-voice-url>\n<wechat-robot-file-url>...</wechat-robot-file-url>\n<wechat-robot-appmsg type=\"数字\">...</wechat-robot-appmsg>\n如果一次返回多个这类标签，必须完整保留每一个标签及其内部内容；如果还有普通文本，可以与这些标签一起返回，但标签本身必须保持完全不变。"

	// 群聊上下文注入：独立 system 消息置于主 system prompt 之后、对话历史之前
	// 这样主 system prompt 部分可最大程度命中前缀缓存
	var prefixMessages []openai.ChatCompletionMessage
	prefixMessages = append(prefixMessages, systemMessage)
	if chatRoomID != "" {
		if groupCtx := s.buildGroupChatContext(chatRoomID, contactWxID, robotCtx.RobotWxID); groupCtx != "" {
			prefixMessages = append(prefixMessages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: groupCtx,
			})
		}
	}
	aiMessages = append(prefixMessages, aiMessages...)

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

// buildGroupChatContext 构建群聊上下文：当前用户元信息 + 最近其他群友消息
func (s *AIChatService) buildGroupChatContext(chatRoomID, senderWxID, robotWxID string) string {
	var sb strings.Builder

	crmRepo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
	member, err := crmRepo.GetChatRoomMember(chatRoomID, senderWxID)
	if err != nil {
		log.Printf("[GroupContext] 获取群成员信息失败: %v", err)
	}
	if member != nil {
		sb.WriteString("[当前对话用户信息]\n")
		if member.Nickname != "" {
			fmt.Fprintf(&sb, "昵称: %s\n", member.Nickname)
		}
		if member.Remark != "" {
			fmt.Fprintf(&sb, "备注: %s\n", member.Remark)
		}
		if member.Avatar != "" {
			fmt.Fprintf(&sb, "头像: %s\n", member.Avatar)
		}
	}

	msgRepo := repository.NewMessageRepo(s.ctx, vars.DB)
	recentMsgs, err := msgRepo.GetRecentChatRoomMessages(chatRoomID, []string{senderWxID, robotWxID}, 10)
	if err != nil {
		log.Printf("[GroupContext] 获取最近群消息失败: %v", err)
	}
	if len(recentMsgs) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("[最近群聊消息]\n")
		for _, msg := range recentMsgs {
			nickname := msg.SenderNickname
			if nickname == "" {
				nickname = msg.SenderWxID
			}
			fmt.Fprintf(&sb, "[%s] %s\n", nickname, msg.Content)
		}
	}

	return sb.String()
}
