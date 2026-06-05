package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"wechat-robot-client/interface/settings"
	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
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

func (s *AIChatService) Chat(robotCtx robotctx.RobotContext, aiMessages []openai.ChatCompletionMessageParamUnion) (openai.ChatCompletionMessage, error) {
	// 获取 AI 配置
	aiConfig := s.config.GetAIConfig()

	// 构建系统提示词
	var basePrompt strings.Builder
	basePrompt.WriteString(aiConfig.Prompt)

	// 注入当前世界时间
	now := time.Now()
	weekdayMap := map[time.Weekday]string{
		time.Sunday:    "星期日",
		time.Monday:    "星期一",
		time.Tuesday:   "星期二",
		time.Wednesday: "星期三",
		time.Thursday:  "星期四",
		time.Friday:    "星期五",
		time.Saturday:  "星期六",
	}
	basePrompt.WriteString("\n\n【当前世界时间】\n")
	fmt.Fprintf(&basePrompt, "%d 年 %d 月 %d 日，%s", now.Year(), int(now.Month()), now.Day(), weekdayMap[now.Weekday()])

	basePrompt.WriteString("\n\n**【特别重要】**如果外部工具返回以下结构化标签，你必须原样逐字返回，不能总结、解释、改写、翻译、补充代码块，也不能省略、合并或调整顺序：\n<wechat-robot-text>...</wechat-robot-text>\n<wechat-robot-image-url>...</wechat-robot-image-url>\n<wechat-robot-video-url>...</wechat-robot-video-url>\n<wechat-robot-voice-url>...</wechat-robot-voice-url>\n<wechat-robot-file-url>...</wechat-robot-file-url>\n<wechat-robot-appmsg type=\"数字\">...</wechat-robot-appmsg>\n如果一次返回多个这类标签，必须完整保留每一个标签及其内部内容；如果还有普通文本，可以与这些标签一起返回，但标签本身必须保持完全不变。")
	if aiConfig.MaxCompletionTokens > 0 {
		fmt.Fprintf(&basePrompt, "\n\n请注意，每次回答不能超过%d个汉字。", aiConfig.MaxCompletionTokens)
	}

	// 构建系统消息
	var systemMessages []openai.ChatCompletionMessageParamUnion
	// 系统提示词
	systemMessages = append(systemMessages, openai.SystemMessage(basePrompt.String()))
	if strings.Contains(robotCtx.FromWxID, "@chatroom") {
		start := time.Now()
		// 群聊上下文：当前用户元信息 + 最近其他群友消息
		if groupCtx := s.buildGroupChatContext(robotCtx.FromWxID, robotCtx.SenderWxID); groupCtx != "" {
			systemMessages = append(systemMessages, openai.SystemMessage(groupCtx))
		}
		log.Printf("[GroupContext] 构建群聊上下文耗时: %v", time.Since(start))
	}
	// 群友单独的对话记录
	aiMessages = append(systemMessages, aiMessages...)

	client := openai.NewClient(
		option.WithAPIKey(aiConfig.APIKey),
		option.WithBaseURL(aiConfig.BaseURL),
	)
	req := openai.ChatCompletionNewParams{
		Model:    aiConfig.Model,
		Messages: aiMessages,
	}

	aiStart := time.Now()
	reply, err := vars.Agent.ChatWithTools(&robotCtx, &client, req)
	log.Printf("[AI] 接口调用耗时: %v", time.Since(aiStart))

	return reply, err
}

func (s *AIChatService) latestChatMessageText(messages []openai.ChatCompletionMessageParamUnion) string {
	for i := len(messages) - 1; i >= 0; i-- {
		text := s.chatMessageParamText(messages[i])
		if strings.TrimSpace(text) != "" {
			return text
		}
	}
	return ""
}

func (s *AIChatService) chatMessageParamText(message openai.ChatCompletionMessageParamUnion) string {
	switch content := message.GetContent().AsAny().(type) {
	case *string:
		return *content
	case *[]openai.ChatCompletionContentPartTextParam:
		var builder strings.Builder
		for _, part := range *content {
			builder.WriteString(part.Text)
		}
		return builder.String()
	case *[]openai.ChatCompletionContentPartUnionParam:
		var builder strings.Builder
		for _, part := range *content {
			if text := part.GetText(); text != nil {
				builder.WriteString(*text)
			}
		}
		return builder.String()
	case *[]openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion:
		var builder strings.Builder
		for _, part := range *content {
			if text := part.GetText(); text != nil {
				builder.WriteString(*text)
			}
		}
		return builder.String()
	default:
		return ""
	}
}

// buildGroupChatContext 构建群聊上下文：当前用户元信息 + 最近其他群友消息
func (s *AIChatService) buildGroupChatContext(chatRoomID, senderWxID string) string {
	var sb strings.Builder

	crmRepo := repository.NewChatRoomMemberRepo(s.ctx, vars.DB)
	member, err := crmRepo.GetChatRoomMember(chatRoomID, senderWxID)
	if err != nil {
		log.Printf("[GroupContext] 获取群成员信息失败: %v", err)
	}
	if member != nil {
		// 性别缺失时通过接口获取并回写数据库
		if member.Sex == nil {
			if detailResp, err := vars.RobotRuntime.GetContactDetail("", []string{senderWxID}); err != nil {
				log.Printf("[GroupContext] 获取联系人详情失败: %v", err)
			} else if len(detailResp.ContactList) > 0 {
				sexVal := detailResp.ContactList[0].Sex
				member.Sex = &sexVal
				if err := crmRepo.UpdateMemberInfo(chatRoomID, senderWxID, map[string]any{"sex": sexVal}); err != nil {
					log.Printf("[GroupContext] 回写性别失败: %v", err)
				}
			}
		}

		sb.WriteString("[当前对话用户信息]\n")
		fmt.Fprintf(&sb, "微信 ID: %s\n", member.WechatID)
		if member.Nickname != "" {
			fmt.Fprintf(&sb, "昵称: %s\n", member.Nickname)
		}
		if member.Remark != "" {
			fmt.Fprintf(&sb, "备注: %s\n", member.Remark)
		}
		if member.Sex != nil {
			gender := "未知"
			switch *member.Sex {
			case 1:
				gender = "男"
			case 2:
				gender = "女"
			}
			fmt.Fprintf(&sb, "性别: %s\n", gender)
		}
		if member.Avatar != "" {
			fmt.Fprintf(&sb, "头像: %s\n", member.Avatar)
		}
	}

	msgRepo := repository.NewMessageRepo(s.ctx, vars.DB)
	excludeWxIDs := make([]string, 0, 2)
	if senderWxID != "" {
		excludeWxIDs = append(excludeWxIDs, senderWxID)
	}
	if robotWxID := vars.RobotRuntime.WxID; robotWxID != "" && robotWxID != senderWxID {
		excludeWxIDs = append(excludeWxIDs, robotWxID)
	}
	recentMsgs, err := msgRepo.GetRecentChatRoomMessages(chatRoomID, excludeWxIDs, 10)
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
			fmt.Fprintf(&sb, "[%s]: %s\n", nickname, msg.Content)
		}
	}

	return sb.String()
}
