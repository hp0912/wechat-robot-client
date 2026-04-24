package openaitools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sashabaranov/go-openai"

	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/vars"
)

type SendRemoteImageTool struct{}

func NewSendRemoteImageTool(knowledgeService plugin.MessageServiceIface) OpenAITool {
	return &SendRemoteImageTool{}

}

func (t *SendRemoteImageTool) GetOpenAITool(robotCtx *robotctx.RobotContext) *openai.Tool {
	systemPrompt, err := t.BuildSystemPrompt(context.Background(), robotCtx)
	if err != nil {
		fmt.Printf("构建系统提示词失败: %v\n", err)
		return nil
	}
	if systemPrompt == "" {
		return nil
	}
	return &openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "send_remote_image",
			Description: "发送远程图片",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"image_url": map[string]string{
						"type":        "string",
						"description": "远程图片的URL地址",
					},
				},
				"required": []string{"image_url"},
			},
		},
	}
}

func (t *SendRemoteImageTool) BuildSystemPrompt(ctx context.Context, robotCtx *robotctx.RobotContext) (string, error) {
	return "发送远程图片", nil
}

func (t *SendRemoteImageTool) ExecuteToolCall(ctx context.Context, robotCtx *robotctx.RobotContext, toolCall openai.ToolCall) (string, bool, error) {
	var args struct {
		ImageURL string `json:"image_url"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("解析参数失败: %w", err)
	}
	if args.ImageURL == "" {
		return "", false, fmt.Errorf("参数 image_url 不能为空")
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	httpResp, err := resty.New().R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(map[string]any{
			"to_wxid":    robotCtx.FromWxID,
			"image_urls": []string{args.ImageURL},
		}).
		SetResult(&result).
		Post(fmt.Sprintf("http://127.0.0.1:%d/api/v1/robot/message/send/image/url", vars.WechatClientPort))
	if err != nil {
		return "", false, fmt.Errorf("发送请求失败: %w", err)
	}
	if httpResp.IsError() {
		return "", false, fmt.Errorf("请求返回错误状态码: %d", httpResp.StatusCode())
	}
	if result.Code != 200 {
		return "", false, fmt.Errorf("接口返回错误: %s", result.Message)
	}
	return "发送成功", false, nil
}
