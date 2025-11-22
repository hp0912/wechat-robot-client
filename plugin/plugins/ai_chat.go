package plugins

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/mcp"
	"wechat-robot-client/service"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
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

func (p *AIChatPlugin) PreAction(ctx *plugin.MessageContext) bool {
	if ctx.ReferMessage != nil {
		if ctx.ReferMessage.Type == model.MsgTypeImage {
			imageUpload := NewAIImageUploadPlugin()
			result := imageUpload.Run(ctx)
			if !result {
				return false
			}

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

func (p *AIChatPlugin) Run(ctx *plugin.MessageContext) bool {
	if !p.PreAction(ctx) {
		return true
	}

	aiTriggerWord := ctx.Settings.GetAITriggerWord()
	aiMessages, err := ctx.MessageService.GetAIMessageContext(ctx.Message)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return true
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
	aiReply, err := aiChatService.Chat(mcp.RobotContext{
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
		return true
	}
	var aiReplyText string
	if aiReply.Content != "" {
		aiReplyText = aiReply.Content
	} else if len(aiReply.MultiContent) > 0 {
		aiReplyText = aiReply.MultiContent[0].Text
	}
	if aiReplyText == "" {
		aiReplyText = "AI返回了空内容。"
	}

	// 检测是否是 MCP 工具调用结果
	if strings.HasPrefix(aiReplyText, "{") {
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
				}
			case ActionTypeSendImageMessage:
				for _, imageURL := range callToolResult.AttachmentURLList {
					err := ctx.MessageService.SendImageMessageByRemoteURL(ctx.Message.FromWxID, imageURL)
					if err != nil {
						log.Println("发送图片消息失败: ", err.Error())
					}
				}
			case ActionTypeSendVideoMessage:
				for _, videoURL := range callToolResult.AttachmentURLList {
					err := ctx.MessageService.SendVideoMessageByRemoteURL(ctx.Message.FromWxID, videoURL)
					if err != nil {
						log.Println("发送视频消息失败: ", err.Error())
					}
				}
			case ActionTypeSendVoiceMessage:
				audioData, err := base64.StdEncoding.DecodeString(callToolResult.Text)
				if err != nil {
					p.SendMessage(ctx, fmt.Sprintf("视频数据解码失败: %v", err))
				}
				audioReader := bytes.NewReader(audioData)
				ctx.MessageService.MsgSendVoice(ctx.Message.FromWxID, audioReader, fmt.Sprintf(".%s", callToolResult.VoiceEncoding))
			case ActionTypeSendAppMessage:
				err := ctx.MessageService.SendAppMessage(ctx.Message.FromWxID, callToolResult.AppType, callToolResult.AppXML)
				if err != nil {
					p.SendMessage(ctx, err.Error())
				}
			case ActionTypeJoinChatRoom:
				err := service.NewChatRoomService(context.Background()).AutoInviteChatRoomMember(callToolResult.Text, []string{ctx.Message.FromWxID})
				if err != nil {
					p.SendMessage(ctx, err.Error())
				}
			default:
				p.SendMessage(ctx, "暂不支持的操作类型。")
			}
			return true
		}
	}

	p.SendMessage(ctx, aiReplyText)

	return true
}
