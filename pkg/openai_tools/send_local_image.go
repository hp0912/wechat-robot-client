package openaitools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go/v3"

	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/vars"
)

type SendLocalImageTool struct{}

func NewSendLocalImageTool() OpenAITool {
	return &SendLocalImageTool{}

}

func (t *SendLocalImageTool) GetOpenAITool(robotCtx *robotctx.RobotContext) *openai.ChatCompletionToolUnionParam {
	systemPrompt, err := t.BuildSystemPrompt(context.Background(), robotCtx)
	if err != nil {
		fmt.Printf("构建系统提示词失败: %v\n", err)
		return nil
	}
	if systemPrompt == "" {
		return nil
	}
	tool := openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "send_local_image",
		Description: openai.String("发送本地图片"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"image_path": map[string]string{
					"type":        "string",
					"description": "本地图片的绝对路径",
				},
			},
			"required": []string{"image_path"},
		},
	})
	return &tool
}

func (t *SendLocalImageTool) BuildSystemPrompt(ctx context.Context, robotCtx *robotctx.RobotContext) (string, error) {
	return "发送本地图片", nil
}

func (t *SendLocalImageTool) ExecuteToolCall(ctx context.Context, robotCtx *robotctx.RobotContext, toolCall openai.ChatCompletionMessageToolCallUnion) (string, bool, error) {
	var args struct {
		ImagePath string `json:"image_path"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("解析参数失败: %w", err)
	}
	if args.ImagePath == "" {
		return "", false, fmt.Errorf("参数 image_path 不能为空")
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	httpResp, err := resty.New().R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(map[string]string{
			"to_wxid":    robotCtx.FromWxID,
			"image_path": args.ImagePath,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("http://127.0.0.1:%s/api/v1/robot/message/send/image/local", vars.WechatClientPort))
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
