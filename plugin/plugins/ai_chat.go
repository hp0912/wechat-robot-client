package plugins

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/openai/openai-go/v3"

	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/service"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"
)

type ActionType int

const (
	ActionTypeSendTextMessage     ActionType = 1   // 发送普通文本消息
	ActionTypeSendLongTextMessage ActionType = 2   // 发送长文本消息
	ActionTypeSendImageMessage    ActionType = 3   // 发送图片消息
	ActionTypeSendVideoMessage    ActionType = 4   // 发送视频消息
	ActionTypeSendAttachMessage   ActionType = 5   // 发送附件消息
	ActionTypeSendVoiceMessage    ActionType = 6   // 发送语音消息
	ActionTypeSendAppMessage      ActionType = 7   // 发送应用消息
	ActionTypeSendEmoticonMessage ActionType = 8   // 发送表情消息
	ActionTypeJoinChatRoom        ActionType = 100 // 加入群聊
	ActionTypeEmoji               ActionType = 147 // 提取表情包
)

type CallToolResult struct {
	IsCallToolResult  bool       `json:"is_call_tool_result,omitempty" jsonschema:"是否为调用工具结果"`
	ActionType        ActionType `json:"action_type" jsonschema:"操作类型: 1-发送普通文本消息, 2-发送长文本消息, 3-发送图片消息, 4-发送视频消息, 5-发送附件消息, 6-发送语音消息, 7-发送应用消息, 8-发送表情消息"`
	Text              string     `json:"text,omitempty" jsonschema:"文本消息内容"`
	AppType           int        `json:"app_type,omitempty" jsonschema:"应用消息类型"`
	AppXML            string     `json:"app_xml,omitempty" jsonschema:"应用消息的XML内容"`
	VoiceEncoding     string     `json:"voice_encoding,omitempty" jsonschema:"语音消息的编码格式"`
	AttachmentURLList []string   `json:"attachment_url_list,omitempty" jsonschema:"附件消息的URL"`
}

type structuredReplyType string

const (
	structuredReplyTypeText   structuredReplyType = "text"
	structuredReplyTypeImage  structuredReplyType = "image"
	structuredReplyTypeVideo  structuredReplyType = "video"
	structuredReplyTypeVoice  structuredReplyType = "voice"
	structuredReplyTypeFile   structuredReplyType = "file"
	structuredReplyTypeAppMsg structuredReplyType = "appmsg"
)

type structuredReplyBlock struct {
	Start   int
	End     int
	Type    structuredReplyType
	Content string
	AppType int
}

type structuredReplyPattern struct {
	Type    structuredReplyType
	Pattern *regexp.Regexp
}

var thinkTagRegexp = regexp.MustCompile(`(?s)<think>.*?</think>|<thinking>.*?</thinking>`)

type AIChatPlugin struct{}

func NewAIChatPlugin() plugin.MessageHandler {
	return &AIChatPlugin{}
}

func (p *AIChatPlugin) GetName() string {
	return "AIChat"
}

func (p *AIChatPlugin) GetLabels() []string {
	return []string{"text", "internal", "chat"}
}

func (p *AIChatPlugin) Match(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AIChatPlugin) PreAction(ctx *plugin.MessageContext) bool {
	if ctx.ReferMessage != nil {
		if ctx.ReferMessage.Type == model.MsgTypeImage {
			imageUpload := NewAIImageUploadPlugin()
			match := imageUpload.Match(ctx)
			if !match {
				return false
			}
			imageUpload.Run(ctx)
			err := ctx.MessageService.SetMessageIsInContext(ctx.ReferMessage)
			if err != nil {
				log.Printf("更新消息上下文失败: %v", err)
				return false
			}
		}
		if ctx.ReferMessage.Type == model.MsgTypeEmoticon {
			emojiUpload := NewAIEmojiUploadPlugin()
			match := emojiUpload.Match(ctx)
			if !match {
				return false
			}
			emojiUpload.Run(ctx)
			err := ctx.MessageService.SetMessageIsInContext(ctx.ReferMessage)
			if err != nil {
				log.Printf("更新消息上下文失败: %v", err)
				return false
			}
		}
		if ctx.ReferMessage.Type == model.MsgTypeVoice {
			voiceUpload := NewAIVoiceUploadPlugin()
			match := voiceUpload.Match(ctx)
			if !match {
				return false
			}
			voiceUpload.Run(ctx)
			err := ctx.MessageService.SetMessageIsInContext(ctx.ReferMessage)
			if err != nil {
				log.Printf("更新消息上下文失败: %v", err)
				return false
			}
		}
		if ctx.ReferMessage.Type == model.MsgTypeVideo {
			videoUpload := NewAIVideoUploadPlugin()
			match := videoUpload.Match(ctx)
			if !match {
				return false
			}
			videoUpload.Run(ctx)
			err := ctx.MessageService.SetMessageIsInContext(ctx.ReferMessage)
			if err != nil {
				log.Printf("更新消息上下文失败: %v", err)
				return false
			}
		}
		if ctx.ReferMessage.Type == model.MsgTypeApp && ctx.ReferMessage.AppMsgType == model.AppMsgTypeAttach {
			attachUpload := NewAIAttachUploadPlugin()
			match := attachUpload.Match(ctx)
			if !match {
				return false
			}
			attachUpload.Run(ctx)
			err := ctx.MessageService.SetMessageIsInContext(ctx.ReferMessage)
			if err != nil {
				log.Printf("更新消息上下文失败: %v", err)
				return false
			}
		}
	}
	return true
}

func (p *AIChatPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AIChatPlugin) SendMessage(ctx *plugin.MessageContext, aiReplyText string) {
	if aiReplyText == "" {
		return
	}
	if ctx.Message.IsChatRoom {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReplyText, ctx.Message.SenderWxID)
	} else {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReplyText)
	}
}

func (p *AIChatPlugin) setChatMessageTextContent(message *openai.ChatCompletionMessageParamUnion, text string) {
	switch {
	case message.OfDeveloper != nil:
		message.OfDeveloper.Content = openai.ChatCompletionDeveloperMessageParamContentUnion{OfString: openai.String(text)}
	case message.OfSystem != nil:
		message.OfSystem.Content = openai.ChatCompletionSystemMessageParamContentUnion{OfString: openai.String(text)}
	case message.OfUser != nil:
		message.OfUser.Content = openai.ChatCompletionUserMessageParamContentUnion{OfString: openai.String(text)}
	case message.OfAssistant != nil:
		message.OfAssistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{OfString: openai.String(text)}
	case message.OfTool != nil:
		message.OfTool.Content = openai.ChatCompletionToolMessageParamContentUnion{OfString: openai.String(text)}
	case message.OfFunction != nil:
		message.OfFunction.Content = openai.String(text)
	default:
		*message = openai.UserMessage(text)
	}
}

func (p *AIChatPlugin) trimAITriggerFromText(text, aiTriggerWord string) string {
	trimmedText := utils.TrimAITriggerAll(text, aiTriggerWord)
	if trimmedText == "" {
		return "在吗？"
	}
	return trimmedText
}

func (p *AIChatPlugin) trimAITriggerFromTextParts(parts []openai.ChatCompletionContentPartTextParam, aiTriggerWord string) {
	for index := range parts {
		parts[index].Text = p.trimAITriggerFromText(parts[index].Text, aiTriggerWord)
	}
}

func (p *AIChatPlugin) trimAITriggerFromUserContentParts(parts []openai.ChatCompletionContentPartUnionParam, aiTriggerWord string) {
	for index := range parts {
		if text := parts[index].GetText(); text != nil {
			*text = p.trimAITriggerFromText(*text, aiTriggerWord)
		}
	}
}

func (p *AIChatPlugin) trimAITriggerFromAssistantContentParts(parts []openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion, aiTriggerWord string) {
	for index := range parts {
		if text := parts[index].GetText(); text != nil {
			*text = p.trimAITriggerFromText(*text, aiTriggerWord)
		}
	}
}

func (p *AIChatPlugin) trimAITriggerFromChatMessage(message *openai.ChatCompletionMessageParamUnion, aiTriggerWord string) {
	switch content := message.GetContent().AsAny().(type) {
	case *string:
		*content = p.trimAITriggerFromText(*content, aiTriggerWord)
	case *[]openai.ChatCompletionContentPartTextParam:
		if len(*content) == 0 {
			p.setChatMessageTextContent(message, "在吗？")
			return
		}
		p.trimAITriggerFromTextParts(*content, aiTriggerWord)
	case *[]openai.ChatCompletionContentPartUnionParam:
		if len(*content) == 0 {
			p.setChatMessageTextContent(message, "在吗？")
			return
		}
		p.trimAITriggerFromUserContentParts(*content, aiTriggerWord)
	case *[]openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion:
		if len(*content) == 0 {
			p.setChatMessageTextContent(message, "在吗？")
			return
		}
		p.trimAITriggerFromAssistantContentParts(*content, aiTriggerWord)
	default:
		p.setChatMessageTextContent(message, "在吗？")
	}
}

func (p *AIChatPlugin) chatMessageText(message openai.ChatCompletionMessageParamUnion) string {
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

func (p *AIChatPlugin) Run(ctx *plugin.MessageContext) {
	if !p.PreAction(ctx) {
		return
	}

	aiTriggerWord := ctx.Settings.GetAITriggerWord()
	aiMessages, err := ctx.MessageService.GetAIMessageContext(ctx.Message)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return
	}
	if ctx.Message.IsChatRoom {
		for index := range aiMessages {
			p.trimAITriggerFromChatMessage(&aiMessages[index], aiTriggerWord)
		}
	}
	aiChatService := service.NewAIChatService(ctx.Context, ctx.Settings)
	var refMessageID int64
	if ctx.ReferMessage != nil {
		refMessageID = ctx.ReferMessage.ID
	}
	aiReply, err := aiChatService.Chat(robotctx.RobotContext{
		WeChatClientPort: vars.WechatClientPort,
		RobotID:          vars.RobotRuntime.RobotID,
		RobotCode:        vars.RobotRuntime.RobotCode,
		DBHost:           vars.MysqlSettings.Host,
		DBPort:           vars.MysqlSettings.Port,
		DBUser:           vars.MysqlSettings.PrivateUser, // 给工具专用的数据库用户，只针对当前机器人数据库有操作权限
		DBPassword:       vars.MysqlSettings.PrivatePassword,
		RobotRedisDB:     vars.RobotRuntime.RobotRedisDB,
		RobotWxID:        vars.RobotRuntime.WxID,
		FromWxID:         ctx.Message.FromWxID,
		SenderWxID:       ctx.Message.SenderWxID,
		MessageID:        ctx.Message.ID,
		RefMessageID:     refMessageID,
	}, aiMessages)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return
	}
	var aiReplyText string
	if aiReply.Content != "" {
		aiReplyText = aiReply.Content
	}
	// aiReplyText 可能包含思维链，<think></think> 标签内的内容是 AI 的思考过程，不应该发送给用户
	aiReplyText = thinkTagRegexp.ReplaceAllString(aiReplyText, "")
	aiReplyText = strings.TrimSpace(aiReplyText)

	if aiReplyText == "" {
		aiReplyText = "AI返回了空内容。"
		if len(aiMessages) > 0 {
			lastMessage := aiMessages[len(aiMessages)-1]
			if strings.Contains(p.chatMessageText(lastMessage), "#调试") {
				debugPayload := map[string]any{
					"ai_messages": aiMessages,
					"ai_reply":    aiReply,
				}
				debugBytes, err := json.Marshal(debugPayload)
				if err != nil {
					aiReplyText = fmt.Sprintf("调试数据序列化失败: %v", err)
				} else {
					aiReplyText = string(debugBytes)
				}
			}
		} else {
			aiReplyText = "没有获取到 AI 对话上下文，请联系管理员。"
		}
	}

	if aiReplyText == vars.AIEnded {
		_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
		return
	}

	// 检测是否是 MCP 工具调用结果
	if strings.HasPrefix(strings.TrimSpace(aiReplyText), "{") {
		var callToolResult CallToolResult
		err = json.Unmarshal([]byte(aiReplyText), &callToolResult)
		if err == nil && callToolResult.IsCallToolResult {
			switch callToolResult.ActionType {
			case ActionTypeSendTextMessage:
				p.SendMessage(ctx, callToolResult.Text)
			case ActionTypeSendLongTextMessage:
				err := ctx.MessageService.SendLongTextMessage(ctx.Message.FromWxID, callToolResult.Text)
				if err != nil {
					p.SendMessage(ctx, err.Error())
				} else {
					_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
				}
			case ActionTypeSendImageMessage:
				for _, imageURL := range callToolResult.AttachmentURLList {
					err := ctx.MessageService.SendImageMessageByRemoteURL(ctx.Message.FromWxID, imageURL)
					if err != nil {
						log.Println("发送图片消息失败: ", err.Error())
					}
				}
				_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
			case ActionTypeSendVideoMessage:
				for _, videoURL := range callToolResult.AttachmentURLList {
					err := ctx.MessageService.SendVideoMessageByRemoteURL(ctx.Message.FromWxID, videoURL)
					if err != nil {
						log.Println("发送视频消息失败: ", err.Error())
					}
				}
				_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
			case ActionTypeSendVoiceMessage:
				audioData, err := base64.StdEncoding.DecodeString(callToolResult.Text)
				if err != nil {
					p.SendMessage(ctx, fmt.Sprintf("视频数据解码失败: %v", err))
				}
				audioReader := bytes.NewReader(audioData)
				err = ctx.MessageService.MsgSendVoice(ctx.Message.FromWxID, audioReader, fmt.Sprintf(".%s", callToolResult.VoiceEncoding))
				if err != nil {
					p.SendMessage(ctx, fmt.Sprintf("发送语音消息失败: %v", err))
				} else {
					_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
				}
			case ActionTypeSendAppMessage:
				err := ctx.MessageService.SendAppMessage(ctx.Message.FromWxID, callToolResult.AppType, callToolResult.AppXML)
				if err != nil {
					p.SendMessage(ctx, err.Error())
				} else {
					_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
				}
			case ActionTypeJoinChatRoom:
				err := service.NewChatRoomService(context.Background()).AutoInviteChatRoomMember(callToolResult.Text, []string{ctx.Message.FromWxID})
				if err != nil {
					p.SendMessage(ctx, err.Error())
				} else {
					_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
				}
			case ActionTypeEmoji:
				imageURL := callToolResult.Text
				if ctx.ReferMessage.AttachmentUrl != "" {
					if strings.HasSuffix(ctx.ReferMessage.AttachmentUrl, "gif") {
						p.SendMessage(ctx, fmt.Sprintf("表情下载地址: %s", ctx.ReferMessage.AttachmentUrl))
					} else {
						err := ctx.MessageService.SendImageMessageByRemoteURL(ctx.Message.FromWxID, ctx.ReferMessage.AttachmentUrl)
						if err != nil {
							log.Println("发送图片消息失败: ", err.Error())
						}
						_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
					}
					return
				}
				ossSettingService := service.NewOSSSettingService(ctx.Context)
				ossSettings, err := ossSettingService.GetOSSSettingService()
				if err != nil {
					p.SendMessage(ctx, "获取 OSS 配置失败，请联系管理员")
					return
				}
				if ossSettings.AutoUploadImage != nil && *ossSettings.AutoUploadImage {
					err := ossSettingService.UploadImageToOSSFromEncryptUrl(ossSettings, ctx.ReferMessage, imageURL)
					if err != nil {
						p.SendMessage(ctx, "上传图片到 OSS 失败，请联系管理员")
						return
					}
					if strings.HasSuffix(ctx.ReferMessage.AttachmentUrl, "gif") {
						p.SendMessage(ctx, fmt.Sprintf("表情下载地址: %s", ctx.ReferMessage.AttachmentUrl))
					} else {
						err := ctx.MessageService.SendImageMessageByRemoteURL(ctx.Message.FromWxID, ctx.ReferMessage.AttachmentUrl)
						if err != nil {
							log.Println("发送图片消息失败: ", err.Error())
						}
						_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
					}
				} else {
					p.SendMessage(ctx, "图片上传未开启，请联系管理员")
				}

			default:
				p.SendMessage(ctx, "暂不支持的操作类型。")
			}
			return
		}
	}

	p.SendMessage(ctx, aiReplyText)
}
