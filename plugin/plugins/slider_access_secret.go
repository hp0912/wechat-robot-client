package plugins

import (
	"strings"

	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/plugin/pkg"
	"wechat-robot-client/vars"
)

type SliderAccessSecretPlugin struct{}

func NewSliderAccessSecretPlugin() plugin.MessageHandler {
	return &SliderAccessSecretPlugin{}
}

func (p *SliderAccessSecretPlugin) GetName() string {
	return "SliderAccessSecret"
}

func (p *SliderAccessSecretPlugin) GetLabels() []string {
	return []string{"text", "chat"}
}

func (p *SliderAccessSecretPlugin) Match(ctx *plugin.MessageContext) bool {
	return strings.HasPrefix(ctx.MessageContent, "#过滑块密钥")
}

func (p *SliderAccessSecretPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return NewChatRoomCommonPlugin().PreAction(ctx)
}

func (p *SliderAccessSecretPlugin) PostAction(ctx *plugin.MessageContext) {
}

func (p *SliderAccessSecretPlugin) Run(ctx *plugin.MessageContext) {
	if !p.PreAction(ctx) {
		return
	}

	if vars.SliderAccessKey == "" {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "滑块访问密钥未配置，请联系管理员", ctx.Message.SenderWxID)
		return
	}

	secret, err := pkg.GenerateSliderAccessSecret(nil, 0)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error(), ctx.Message.SenderWxID)
		return
	}

	ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, secret+"\n\n有效期24小时", ctx.Message.SenderWxID)
}
