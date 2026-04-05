package service

import (
	"context"
	"log"
	"strings"
	"wechat-robot-client/interface/ai"
)

type ToolBindingRegistry struct {
	providers []ai.ToolBindingProvider
}

var _ ai.ToolBindingRegistry = (*ToolBindingRegistry)(nil)

func NewToolBindingRegistry(providers ...ai.ToolBindingProvider) *ToolBindingRegistry {
	return &ToolBindingRegistry{providers: providers}
}

func (r *ToolBindingRegistry) BuildToolBinding(ctx context.Context, toolCtx ai.ToolBindingContext) ai.ToolBinding {
	merged := ai.ToolBinding{}
	for _, provider := range r.providers {
		binding, err := provider.BuildToolBinding(ctx, toolCtx)
		if err != nil {
			log.Printf("[ToolBinding:%s] 构建失败: %v", provider.Name(), err)
			continue
		}
		if binding.Prompt != "" {
			if merged.Prompt == "" {
				merged.Prompt = binding.Prompt
			} else {
				merged.Prompt += "\n" + strings.TrimLeft(binding.Prompt, "\n")
			}
		}
		if len(binding.Tools) > 0 {
			merged.Tools = append(merged.Tools, binding.Tools...)
		}
	}
	return merged
}
