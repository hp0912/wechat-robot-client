package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/openai/openai-go/v3"
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
func (s *AgentService) GetAllTools(robotCtx *robotctx.RobotContext) ([]openai.ChatCompletionToolUnionParam, error) {
	var tools []openai.ChatCompletionToolUnionParam
	// 从内部工具管理器获取工具
	internalTools := s.internalToolsManager.GetOpenAITools(robotCtx)
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
	robotCtx *robotctx.RobotContext,
	client *openai.Client,
	req openai.ChatCompletionNewParams,
) (openai.ChatCompletionMessage, error) {
	if len(req.Messages) == 0 {
		return openai.ChatCompletionMessage{}, fmt.Errorf("messages cannot be empty")
	}

	// 获取所有可用工具
	tools, err := s.GetAllTools(robotCtx)
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("failed to get tools: %w", err)
	}

	// 如果没有可用工具，直接调用AI
	if len(tools) == 0 {
		msg, _, err := s.streamChatCompletion(client, req)
		return msg, err
	}

	req.Tools = tools

	// 构建包含工具描述的系统提示词，追加到首条 system 消息或前置新消息
	toolsPrompt, err := s.BuildSystemPrompt(s.ctx, robotCtx)
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("failed to build system prompt: %w", err)
	}
	if req.Messages[0].OfSystem != nil {
		existing := req.Messages[0].OfSystem.Content.OfString.Value
		req.Messages[0].OfSystem.Content.OfString = openai.String(existing + "\n" + toolsPrompt)
	} else {
		req.Messages = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(toolsPrompt)}, req.Messages...)
	}

	for range vars.MaxToolsIterations {
		// 调用AI
		msg, reasoning, err := s.streamChatCompletion(client, req)
		if err != nil {
			return openai.ChatCompletionMessage{}, fmt.Errorf("failed to call ai: %w", err)
		}

		// 没有工具调用，返回结果
		if len(msg.ToolCalls) == 0 {
			return msg, nil
		}

		asstParam := msg.ToParam()
		if reasoning != "" && asstParam.OfAssistant != nil {
			asstParam.OfAssistant.SetExtraFields(map[string]any{
				"reasoning_content": reasoning,
			})
		}
		req.Messages = append(req.Messages, asstParam)

		// 执行所有工具调用
		for _, tc := range msg.ToolCalls {
			log.Printf("Executing tool: %s", tc.Function.Name)

			var result string
			var immediately bool
			var err error

			if s.skillsManager.IsSkillTool(tc.Function.Name) {
				// skill 工具调用
				result, err = s.skillsManager.ExecuteToolCall(*robotCtx, tc)
				immediately = result == vars.AIEnded || strings.HasSuffix(result, "\n"+vars.AIEnded)
				if immediately {
					result = vars.AIEnded
				}
				if tc.Function.Name == "execute_skill_script" {
					log.Printf("工具[%s]执行结果:\n%s\n", tc.Function.Name, result)
				}
			} else if s.internalToolsManager.IsOpenAITool(tc.Function.Name) {
				// 内部工具调用
				result, immediately, err = s.internalToolsManager.ExecuteToolCall(s.ctx, robotCtx, tc)
			} else {
				// MCP 工具调用
				result, immediately, err = s.mcpManager.ExecuteToolCall(s.ctx, *robotCtx, tc)
			}

			if err == nil {
				// 工具调用结果立即返回
				if immediately {
					return openai.ChatCompletionMessage{Content: result}, nil
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

			// 将工具结果追加到消息历史
			req.Messages = append(req.Messages, openai.ToolMessage(result, tc.ID))
		}
	}

	return openai.ChatCompletionMessage{}, fmt.Errorf("max iterations reached without final answer")
}

// streamChatCompletion 通过流式接口调用 AI 并用 accumulator 汇总完整消息。
// 第二个返回值为累积的 reasoning_content（思考内容），用于回写给后续请求。
func (s *AgentService) streamChatCompletion(
	client *openai.Client,
	req openai.ChatCompletionNewParams,
) (openai.ChatCompletionMessage, string, error) {
	stream := client.Chat.Completions.NewStreaming(s.ctx, req)
	acc := openai.ChatCompletionAccumulator{}
	var reasoningSB strings.Builder
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
		// openai-go v3 SDK 没有 reasoning_content 字段，从 ExtraFields 原始 JSON 中提取
		if len(chunk.Choices) > 0 {
			if rcField, ok := chunk.Choices[0].Delta.JSON.ExtraFields["reasoning_content"]; ok {
				raw := rcField.Raw()
				if raw != "" && raw != "null" {
					var rc string
					if err := json.Unmarshal([]byte(raw), &rc); err == nil {
						reasoningSB.WriteString(rc)
					}
				}
			}
		}
	}
	if err := stream.Err(); err != nil {
		return openai.ChatCompletionMessage{}, "", fmt.Errorf("stream error: %w", err)
	}
	if len(acc.Choices) == 0 {
		return openai.ChatCompletionMessage{}, "", fmt.Errorf("no choices in response")
	}
	msg := acc.Choices[0].Message
	log.Printf("Stream finished with reason: %s, toolCalls: %d, content length: %d, reasoning length: %d",
		acc.Choices[0].FinishReason, len(msg.ToolCalls), len(msg.Content), reasoningSB.Len())
	return msg, reasoningSB.String(), nil
}
