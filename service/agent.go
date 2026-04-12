package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"

	"wechat-robot-client/interface/ai"
	"wechat-robot-client/pkg/mcp"
	openaitools "wechat-robot-client/pkg/openai_tools"
	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/pkg/skills"
	"wechat-robot-client/vars"
)

type AgentService struct {
	ctx                  context.Context
	db                   *gorm.DB
	mcpManager           *mcp.MCPManager
	skillsManager        *skills.SkillsManager
	internalToolsManager *openaitools.OpenAIToolsManager
}

var _ ai.AgentService = (*AgentService)(nil)

func NewAgentService(ctx context.Context, db *gorm.DB, knowledgeService ai.KnowledgeService) *AgentService {
	skillsRepo := NewSkillRepoAdapter(db)
	mcpManager := mcp.NewMCPManager(db)
	skillsManager := skills.NewSkillsManager(vars.SkillsDir, skillsRepo)
	internalToolsManager := openaitools.NewOpenAIToolsManager(db, knowledgeService)

	return &AgentService{
		ctx:                  ctx,
		db:                   db,
		mcpManager:           mcpManager,
		skillsManager:        skillsManager,
		internalToolsManager: internalToolsManager,
	}
}

func (s *AgentService) Name() string {
	return "AgentService"
}

func (s *AgentService) Initialize() error {
	log.Println("Initializing internal Tools Manager...")
	err := s.internalToolsManager.Initialize()
	if err != nil {
		return err
	}
	log.Println("Initializing MCP Manager...")
	err = s.mcpManager.Initialize()
	if err != nil {
		return err
	}
	log.Println("Initializing Skills Manager...")
	return s.skillsManager.Initialize()
}

func (s *AgentService) Shutdown(ctx context.Context) error {
	internalToolsErr := s.internalToolsManager.Shutdown()
	mcpErr := s.mcpManager.Shutdown()
	skillsErr := s.skillsManager.Shutdown()
	if internalToolsErr != nil {
		log.Printf("Error shutting down Internal Tools Manager: %v\n", internalToolsErr)
	}
	if mcpErr != nil {
		log.Printf("Error shutting down MCP Manager: %v\n", mcpErr)
	}
	if skillsErr != nil {
		log.Printf("Error shutting down Skills Manager: %v\n", skillsErr)
	}
	return nil
}

func (s *AgentService) GetMCPManager() *mcp.MCPManager {
	return s.mcpManager
}

func (s *AgentService) GetSkillsManager() *skills.SkillsManager {
	return s.skillsManager
}

// GetAllTools 获取所有可用工具（OpenAI格式）
func (s *AgentService) GetAllTools() ([]openai.Tool, error) {
	var tools []openai.Tool
	// 从内部工具管理器获取工具
	internalTools := s.internalToolsManager.GetOpenAITools()
	tools = append(tools, internalTools...)
	// 从MCP获取工具
	mcpTools, err := s.mcpManager.GetOpenAITools(s.ctx)
	if err != nil {
		return tools, err
	}
	tools = append(tools, mcpTools...)
	// 从Skills获取工具
	skillTools := s.skillsManager.GetOpenAITools()
	tools = append(tools, skillTools...)
	return tools, nil
}

// BuildSystemPrompt 构建包含工具描述的系统提示词
func (s *AgentService) BuildSystemPrompt(ctx context.Context, robotCtx *robotctx.RobotContext) (string, error) {
	var sb strings.Builder

	internalToolsPrompt, err := s.internalToolsManager.BuildSystemPrompt(ctx, robotCtx)
	if err != nil {
		return "", fmt.Errorf("failed to build internal tools system prompt: %w", err)
	}
	sb.WriteString(internalToolsPrompt)
	sb.WriteString("\n\n")

	mcpPrompt, err := s.mcpManager.BuildSystemPrompt(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to build MCP system prompt: %w", err)
	}
	sb.WriteString(mcpPrompt)
	sb.WriteString("\n\n")
	skillsPrompt := s.skillsManager.BuildSystemPrompt()
	sb.WriteString(skillsPrompt)

	return sb.String(), nil
}

func (s *AgentService) ChatWithTools(
	robotCtx robotctx.RobotContext,
	client *openai.Client,
	req openai.ChatCompletionRequest,
) (openai.ChatCompletionMessage, error) {
	if len(req.Messages) == 0 {
		return openai.ChatCompletionMessage{}, fmt.Errorf("messages cannot be empty")
	}
	// 获取所有可用工具
	tools, err := s.GetAllTools()
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("failed to get tools: %w", err)
	}
	// 如果没有可用工具，直接调用AI
	if len(tools) == 0 {
		return s.chatWithoutTools(client, req)
	}

	req.Tools = tools

	// 构建包含工具描述的系统提示词
	if req.Messages[0].Role == openai.ChatMessageRoleSystem {
		toolsPrompt, err := s.BuildSystemPrompt(s.ctx, &robotCtx)
		if err != nil {
			return openai.ChatCompletionMessage{}, fmt.Errorf("failed to build system prompt: %w", err)
		}
		req.Messages[0].Content += "\n" + toolsPrompt
	} else {
		toolsPrompt, err := s.BuildSystemPrompt(s.ctx, &robotCtx)
		if err != nil {
			return openai.ChatCompletionMessage{}, fmt.Errorf("failed to build system prompt: %w", err)
		}
		systemMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: toolsPrompt,
		}
		req.Messages = append([]openai.ChatCompletionMessage{systemMsg}, req.Messages...)
	}

	for range vars.MaxToolsIterations {
		// 调用AI
		assistantMsg, err := s.streamChatCompletion(client, req)
		if err != nil {
			return openai.ChatCompletionMessage{}, fmt.Errorf("failed to call ai: %w", err)
		}

		// 检查是否需要调用工具
		if len(assistantMsg.ToolCalls) == 0 {
			// 没有工具调用，返回结果
			return assistantMsg, nil
		}

		// 将assistant消息添加到历史
		req.Messages = append(req.Messages, assistantMsg)

		// 执行所有工具调用
		for _, toolCall := range assistantMsg.ToolCalls {
			log.Printf("Executing tool: %s", toolCall.Function.Name)

			var result string
			var immediately bool
			var err error

			if s.skillsManager.IsSkillTool(toolCall.Function.Name) {
				// skill 工具调用
				result, err = s.skillsManager.ExecuteToolCall(robotCtx, toolCall)
				immediately = result == vars.AIEnded || strings.HasSuffix(result, "\n"+vars.AIEnded)
				if immediately {
					result = vars.AIEnded
				}
				if toolCall.Function.Name == "execute_skill_script" {
					log.Printf("工具[%s]执行结果:\n%s\n", toolCall.Function.Name, result)
				}
			} else if s.internalToolsManager.IsOpenAITool(toolCall.Function.Name) {
				// 内部工具调用
				result, immediately, err = s.internalToolsManager.ExecuteToolCall(s.ctx, &robotCtx, toolCall)
			} else {
				// MCP 工具调用
				result, immediately, err = s.mcpManager.ExecuteToolCall(s.ctx, robotCtx, toolCall)
			}

			if err == nil {
				// 工具调用结果立即返回
				if immediately {
					return openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleAssistant,
						Content:    result,
						ToolCallID: toolCall.ID,
					}, nil
				}
				// 工具返回空结果时，补充默认提示，避免API报错
				if result == "" {
					result = "Tool executed successfully (no output)."
				}
			} else {
				// 工具执行失败，返回错误信息
				result = err.Error()
				log.Println(result)
			}

			// 将工具结果添加到消息历史
			toolResultMsg := openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			}
			req.Messages = append(req.Messages, toolResultMsg)
		}
	}

	// 达到最大迭代次数，返回最后的消息
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		if lastMsg.Role == openai.ChatMessageRoleAssistant {
			return lastMsg, nil
		}
	}

	return openai.ChatCompletionMessage{}, fmt.Errorf("max iterations reached without final answer")
}

// chatWithoutTools 不使用工具的简单聊天
func (s *AgentService) chatWithoutTools(
	client *openai.Client,
	req openai.ChatCompletionRequest,
) (openai.ChatCompletionMessage, error) {
	return s.streamChatCompletion(client, req)
}

// streamChatCompletion 处理流式响应并返回完整消息
func (s *AgentService) streamChatCompletion(
	client *openai.Client,
	req openai.ChatCompletionRequest,
) (openai.ChatCompletionMessage, error) {
	req.Stream = true
	stream, err := client.CreateChatCompletionStream(s.ctx, req)
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	var assistantMsg openai.ChatCompletionMessage
	assistantMsg.Role = openai.ChatMessageRoleAssistant
	var toolCalls []openai.ToolCall
	var finishReason openai.FinishReason

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return openai.ChatCompletionMessage{}, fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) == 0 {
			continue
		}

		choice := response.Choices[0]
		delta := choice.Delta

		if delta.Role != "" {
			assistantMsg.Role = delta.Role
		}

		assistantMsg.Content += delta.Content

		if delta.Refusal != "" {
			assistantMsg.Refusal += delta.Refusal
		}

		for _, tc := range delta.ToolCalls {
			if tc.Index == nil {
				continue
			}
			idx := *tc.Index

			for len(toolCalls) <= idx {
				toolCalls = append(toolCalls, openai.ToolCall{})
			}

			if tc.ID != "" {
				toolCalls[idx].ID = tc.ID
			}
			if tc.Type != "" {
				toolCalls[idx].Type = tc.Type
			}
			if tc.Function.Name != "" {
				toolCalls[idx].Function.Name = tc.Function.Name
			}
			toolCalls[idx].Function.Arguments += tc.Function.Arguments
		}

		// 处理旧版 FunctionCall（向后兼容）
		if delta.FunctionCall != nil {
			if assistantMsg.FunctionCall == nil {
				assistantMsg.FunctionCall = &openai.FunctionCall{}
			}
			if delta.FunctionCall.Name != "" {
				assistantMsg.FunctionCall.Name = delta.FunctionCall.Name
			}
			assistantMsg.FunctionCall.Arguments += delta.FunctionCall.Arguments
		}

		if choice.FinishReason != "" {
			finishReason = choice.FinishReason
		}
	}

	assistantMsg.ToolCalls = toolCalls

	if finishReason != "" {
		log.Printf("Stream finished with reason: %s, toolCalls: %d, content length: %d",
			finishReason, len(toolCalls), len(assistantMsg.Content))
	}

	return assistantMsg, nil
}
