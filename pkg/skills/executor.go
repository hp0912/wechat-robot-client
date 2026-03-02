package skills

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// SkillToolExecutor Skills 工具执行器
// 将 Skills 引擎注册为 OpenAI function tools，在 AI 工具调用循环中执行
type SkillToolExecutor struct {
	manager *Manager
}

// NewSkillToolExecutor 创建 Skills 工具执行器
func NewSkillToolExecutor(manager *Manager) *SkillToolExecutor {
	return &SkillToolExecutor{manager: manager}
}

// ToolNameActivate activate_skill 工具名称
const ToolNameActivate = "activate_skill"

// ToolNameReadResource read_skill_resource 工具名称
const ToolNameReadResource = "read_skill_resource"

// GetOpenAITools 返回 Skills 相关的 OpenAI Tool 定义
func (e *SkillToolExecutor) GetOpenAITools() []openai.Tool {
	summaries := e.manager.GetAllSummaries()
	if len(summaries) == 0 {
		return nil
	}

	// 获取可用 skill 名称列表
	var skillNames []string
	for _, s := range summaries {
		skillNames = append(skillNames, s.Name)
	}

	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        ToolNameActivate,
				Description: "激活一个 Agent Skill，加载其完整的操作指令。当你判断用户任务与某个可用 Skill 相关时调用此工具。激活后你将获得该 Skill 的完整操作指令，请严格按照指令执行任务。",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"skill_name": {
							Type:        jsonschema.String,
							Description: "要激活的 Skill 名称",
							Enum:        skillNames,
						},
					},
					Required: []string{"skill_name"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        ToolNameReadResource,
				Description: "读取已激活 Skill 中的附属资源文件。当 Skill 指令中引用了外部文件（如 scripts/、references/、assets/ 下的文件）时调用此工具。",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"skill_name": {
							Type:        jsonschema.String,
							Description: "Skill 名称",
						},
						"file_path": {
							Type:        jsonschema.String,
							Description: "要读取的文件相对路径，例如 scripts/extract.py 或 references/REFERENCE.md",
						},
					},
					Required: []string{"skill_name", "file_path"},
				},
			},
		},
	}

	return tools
}

// IsSkillTool 判断工具调用是否是 Skills 引擎的工具
func (e *SkillToolExecutor) IsSkillTool(toolName string) bool {
	return toolName == ToolNameActivate || toolName == ToolNameReadResource
}

// ExecuteToolCall 执行 Skills 工具调用，返回结果字符串
func (e *SkillToolExecutor) ExecuteToolCall(toolCall openai.ToolCall) (string, error) {
	switch toolCall.Function.Name {
	case ToolNameActivate:
		return e.executeActivate(toolCall.Function.Arguments)
	case ToolNameReadResource:
		return e.executeReadResource(toolCall.Function.Arguments)
	default:
		return "", fmt.Errorf("unknown skill tool: %s", toolCall.Function.Name)
	}
}

// executeActivate 执行 activate_skill
func (e *SkillToolExecutor) executeActivate(argsJSON string) (string, error) {
	var args struct {
		SkillName string `json:"skill_name"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse activate_skill args: %w", err)
	}

	skill, err := e.manager.ActivateSkill(args.SkillName)
	if err != nil {
		return "", err
	}

	log.Printf("[Skills] Activated skill: %s", args.SkillName)

	// 返回完整的 Skill 指令
	result := fmt.Sprintf(`# Skill「%s」已激活

以下是该 Skill 的完整操作指令，请严格遵循执行：

---

%s

---

如需读取 Skill 中引用的附属文件，请调用 read_skill_resource 工具，传入 skill_name="%s" 和对应的 file_path。`,
		skill.Name, skill.Instructions, skill.Name)

	return result, nil
}

// executeReadResource 执行 read_skill_resource
func (e *SkillToolExecutor) executeReadResource(argsJSON string) (string, error) {
	var args struct {
		SkillName string `json:"skill_name"`
		FilePath  string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse read_skill_resource args: %w", err)
	}

	content, err := e.manager.ReadResource(args.SkillName, args.FilePath)
	if err != nil {
		return "", err
	}

	log.Printf("[Skills] Read resource: %s / %s (%d bytes)", args.SkillName, args.FilePath, len(content))
	return content, nil
}
