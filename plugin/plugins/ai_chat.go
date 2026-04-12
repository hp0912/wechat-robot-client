package plugins

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"

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

var thinkTagRegexp = regexp.MustCompile(`(?s)<think>.*?</think>`)
var unclosedThinkTagRegexp = regexp.MustCompile(`(?s)<think>.*$`)

var structuredReplyPatterns = []structuredReplyPattern{
	{
		Type:    structuredReplyTypeText,
		Pattern: regexp.MustCompile(`(?s)<wechat-robot-text>(.*?)</wechat-robot-text>`),
	},
	{
		Type:    structuredReplyTypeImage,
		Pattern: regexp.MustCompile(`(?s)<wechat-robot-image-url>(.*?)</wechat-robot-image-url>`),
	},
	{
		Type:    structuredReplyTypeVideo,
		Pattern: regexp.MustCompile(`(?s)<wechat-robot-video-url>(.*?)</wechat-robot-video-url>`),
	},
	{
		Type:    structuredReplyTypeVoice,
		Pattern: regexp.MustCompile(`(?s)<wechat-robot-voice-url>(.*?)</wechat-robot-voice-url>`),
	},
	{
		Type:    structuredReplyTypeFile,
		Pattern: regexp.MustCompile(`(?s)<wechat-robot-file-url>(.*?)</wechat-robot-file-url>`),
	},
	{
		Type:    structuredReplyTypeAppMsg,
		Pattern: regexp.MustCompile(`(?s)<wechat-robot-appmsg\s+type=['"](\d+)['"]>(.*?)</wechat-robot-appmsg>`),
	},
}

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

func isRemoteStructuredReplyContent(content string) bool {
	trimmedContent := strings.ToLower(strings.TrimSpace(content))
	return strings.HasPrefix(trimmedContent, "http://") || strings.HasPrefix(trimmedContent, "https://")
}

func extractStructuredReplyBlocks(aiReplyText string) ([]structuredReplyBlock, string) {
	var blocks []structuredReplyBlock
	for _, pattern := range structuredReplyPatterns {
		matches := pattern.Pattern.FindAllStringSubmatchIndex(aiReplyText, -1)
		for _, match := range matches {
			if len(match) < 4 {
				continue
			}
			block := structuredReplyBlock{
				Start:   match[0],
				End:     match[1],
				Type:    pattern.Type,
				Content: strings.TrimSpace(aiReplyText[match[2]:match[3]]),
			}
			if pattern.Type == structuredReplyTypeAppMsg {
				if len(match) < 6 {
					continue
				}
				appType, err := strconv.Atoi(aiReplyText[match[2]:match[3]])
				if err != nil {
					continue
				}
				block.AppType = appType
				block.Content = strings.TrimSpace(aiReplyText[match[4]:match[5]])
			}
			blocks = append(blocks, block)
		}
	}

	if len(blocks) == 0 {
		return nil, aiReplyText
	}

	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i].Start == blocks[j].Start {
			return blocks[i].End < blocks[j].End
		}
		return blocks[i].Start < blocks[j].Start
	})

	filteredBlocks := make([]structuredReplyBlock, 0, len(blocks))
	var remainingText strings.Builder
	current := 0
	for _, block := range blocks {
		if block.Start < current {
			continue
		}
		remainingText.WriteString(aiReplyText[current:block.Start])
		current = block.End
		filteredBlocks = append(filteredBlocks, block)
	}
	remainingText.WriteString(aiReplyText[current:])

	return filteredBlocks, remainingText.String()
}

func (p *AIChatPlugin) handleStructuredReplyBlocks(ctx *plugin.MessageContext, aiReplyText string) bool {
	blocks, remainingText := extractStructuredReplyBlocks(aiReplyText)
	if len(blocks) == 0 {
		return false
	}

	for _, block := range blocks {
		switch block.Type {
		case structuredReplyTypeText:
			multiContentText := strings.TrimSpace(block.Content)
			if multiContentText != "" {
				p.SendMessage(ctx, multiContentText)
			}
		case structuredReplyTypeImage:
			if isRemoteStructuredReplyContent(block.Content) {
				if err := ctx.MessageService.SendImageMessageByRemoteURL(ctx.Message.FromWxID, block.Content); err != nil {
					log.Println("发送结构化图片消息失败: ", err.Error())
				}
				continue
			}
			if err := ctx.MessageService.SendImageMessageByLocalPath(ctx.Message.FromWxID, block.Content); err != nil {
				log.Println("发送结构化图片消息失败: ", err.Error())
			}
			_ = os.Remove(block.Content)
		case structuredReplyTypeVideo:
			if isRemoteStructuredReplyContent(block.Content) {
				if err := ctx.MessageService.SendVideoMessageByRemoteURL(ctx.Message.FromWxID, block.Content); err != nil {
					log.Println("发送结构化视频消息失败: ", err.Error())
				}
				continue
			}
			if err := ctx.MessageService.SendVideoMessageByLocalPath(ctx.Message.FromWxID, block.Content); err != nil {
				log.Println("发送结构化视频消息失败: ", err.Error())
			}
			_ = os.Remove(block.Content)
		case structuredReplyTypeVoice:
			if isRemoteStructuredReplyContent(block.Content) {
				log.Println("结构化语音消息暂不支持远程路径: ", block.Content)
				continue
			}
			if err := ctx.MessageService.SendVoiceMessageByLocalPath(ctx.Message.FromWxID, block.Content); err != nil {
				log.Println("发送结构化语音消息失败: ", err.Error())
			}
			_ = os.Remove(block.Content)
		case structuredReplyTypeFile:
			if isRemoteStructuredReplyContent(block.Content) {
				log.Println("结构化文件消息暂不支持远程路径: ", block.Content)
				continue
			}
			if err := ctx.MessageService.SendFileMessageByLocalPath(ctx.Message.FromWxID, block.Content); err != nil {
				log.Println("发送结构化文件消息失败: ", err.Error())
			}
			_ = os.Remove(block.Content)
		case structuredReplyTypeAppMsg:
			if err := ctx.MessageService.SendAppMessage(ctx.Message.FromWxID, block.AppType, block.Content); err != nil {
				log.Println("发送结构化应用消息失败: ", err.Error())
			}
		}
	}

	if trimmedText := strings.TrimSpace(remainingText); trimmedText != "" {
		// 如果 trimmedText 文本中间有三行或者以上连续的空行，则替换成一行
		trimmedText = regexp.MustCompile(`(?:\r?\n){4,}`).ReplaceAllString(trimmedText, "\n\n")
		p.SendMessage(ctx, trimmedText)
	} else {
		_ = ctx.MessageService.ToolsCompleted(ctx.Message.FromWxID, ctx.Message.SenderWxID)
	}

	return true
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
			// 去除群聊中的AI触发词
			aiMessages[index].Content = utils.TrimAITriggerAll(aiMessages[index].Content, aiTriggerWord)
			if aiMessages[index].Content == "" && len(aiMessages[index].MultiContent) == 0 {
				aiMessages[index].Content = "在吗？"
			}
			for index2 := range aiMessages[index].MultiContent {
				// 去除群聊中的AI触发词
				multiContentText := utils.TrimAITriggerAll(aiMessages[index].MultiContent[index2].Text, aiTriggerWord)
				if multiContentText == "" && aiMessages[index].MultiContent[index2].ImageURL == nil {
					if aiMessages[index].MultiContent[index2].Type == openai.ChatMessagePartTypeText {
						multiContentText = "在吗？"
					} else {
						multiContentText = "链接："
					}
				}
				aiMessages[index].MultiContent[index2].Text = multiContentText
			}
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
	} else if len(aiReply.MultiContent) > 0 {
		aiReplyText = aiReply.MultiContent[0].Text
	}
	// aiReplyText 可能包含思维链，<think></think> 标签内的内容是 AI 的思考过程，不应该发送给用户
	aiReplyText = thinkTagRegexp.ReplaceAllString(aiReplyText, "")
	aiReplyText = strings.TrimSpace(aiReplyText)

	if aiReplyText == "" {
		aiReplyText = "AI返回了空内容。"
		if len(aiMessages) > 0 {
			lastMessage := aiMessages[len(aiMessages)-1]
			if strings.Contains(lastMessage.Content, "#调试") {
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

	if p.handleStructuredReplyBlocks(ctx, aiReplyText) {
		return
	}

	p.SendMessage(ctx, aiReplyText)
}
