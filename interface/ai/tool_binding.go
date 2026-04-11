package ai

import (
	"context"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robotctx"
)

// ToolBinding 表示聊天场景下注入的一组提示词和本地工具。
// 提示词与工具由同一构建逻辑产出，避免出现提示词和实际可用工具不一致。
type ToolBinding struct {
	Prompt string
	Tools  []ExtraTool
}

// ToolBindingContext 描述当前对话场景，用于按场景构建提示词和工具。
type ToolBindingContext struct {
	RobotContext     robotctx.RobotContext
	ChatRoomID       string
	ContactWxID      string
	LastUserQuery    string
	ChatRoomSettings *model.ChatRoomSettings // 预先查询的群聊设置，避免各 Provider 重复查询
}

type ToolBindingProvider interface {
	Name() string
	BuildToolBinding(ctx context.Context, toolCtx ToolBindingContext) (ToolBinding, error)
}

type ToolBindingRegistry interface {
	BuildToolBinding(ctx context.Context, toolCtx ToolBindingContext) ToolBinding
}
