package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"wechat-robot-client/pkg/robotctx"

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

// ToolNameExecuteScript execute_skill_script 工具名称
const ToolNameExecuteScript = "execute_skill_script"

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
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        ToolNameExecuteScript,
				Description: "执行已激活 Skill 中的脚本文件。当 Skill 指令要求运行某个脚本（如 Python/Shell 脚本）时调用此工具。脚本将在 Skill 目录下执行。",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"skill_name": {
							Type:        jsonschema.String,
							Description: "Skill 名称",
						},
						"script_path": {
							Type:        jsonschema.String,
							Description: "要执行的脚本相对路径，例如 scripts/convert.py 或 scripts/build.sh",
						},
						"args": {
							Type:        jsonschema.String,
							Description: "传给脚本的命令行参数（空格分隔），可选",
						},
					},
					Required: []string{"skill_name", "script_path"},
				},
			},
		},
	}

	return tools
}

// IsSkillTool 判断工具调用是否是 Skills 引擎的工具
func (e *SkillToolExecutor) IsSkillTool(toolName string) bool {
	return toolName == ToolNameActivate || toolName == ToolNameReadResource || toolName == ToolNameExecuteScript
}

// ExecuteToolCall 执行 Skills 工具调用，返回结果字符串
func (e *SkillToolExecutor) ExecuteToolCall(robotCtx robotctx.RobotContext, toolCall openai.ToolCall) (string, error) {
	switch toolCall.Function.Name {
	case ToolNameActivate:
		return e.executeActivate(toolCall.Function.Arguments)
	case ToolNameReadResource:
		return e.executeReadResource(toolCall.Function.Arguments)
	case ToolNameExecuteScript:
		return e.executeScript(robotCtx, toolCall.Function.Arguments)
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

// executeScript 执行 execute_skill_script
func (e *SkillToolExecutor) executeScript(robotCtx robotctx.RobotContext, argsJSON string) (string, error) {
	var args struct {
		SkillName  string `json:"skill_name"`
		ScriptPath string `json:"script_path"`
		Args       string `json:"args"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse execute_skill_script args: %w", err)
	}

	e.manager.mu.RLock()
	skill, ok := e.manager.skills[args.SkillName]
	e.manager.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("skill '%s' not found", args.SkillName)
	}

	// 安全检查：防止路径遍历
	cleanPath := filepath.Clean(args.ScriptPath)
	if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("invalid script path: %s", args.ScriptPath)
	}

	absScript := filepath.Join(skill.Path, cleanPath)
	// 确认脚本确实在 Skill 目录内
	if !strings.HasPrefix(absScript, filepath.Clean(skill.Path)+string(filepath.Separator)) {
		return "", fmt.Errorf("script path escapes skill directory: %s", args.ScriptPath)
	}

	// 确定执行器
	var cmdArgs []string
	ext := strings.ToLower(filepath.Ext(cleanPath))
	switch ext {
	case ".py":
		cmdArgs = append(cmdArgs, "python3", absScript)
	case ".sh":
		cmdArgs = append(cmdArgs, "sh", absScript)
	case ".js":
		cmdArgs = append(cmdArgs, "node", absScript)
	default:
		// 尝试直接执行
		cmdArgs = append(cmdArgs, absScript)
	}

	// 追加用户参数
	if args.Args != "" {
		cmdArgs = append(cmdArgs, strings.Fields(args.Args)...)
	}

	log.Printf("[Skills] Executing script: %s (args: %s)", absScript, args.Args)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = skill.Path

	// 注入环境变量（继承进程环境 → Skill 配置变量 → RobotContext）
	env := os.Environ()
	for _, ev := range skill.EnvVars {
		if ev.Key != "" {
			env = append(env, ev.Key+"="+ev.Value)
		}
	}
	env = append(env, robotCtx.ToEnvVars()...)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	result := string(output)

	// 截断过长输出
	const maxLen = 50000
	if len(result) > maxLen {
		result = result[:maxLen] + "\n...(output truncated)"
	}

	if err != nil {
		return fmt.Sprintf("Script execution failed: %v\n\nOutput:\n%s", err, result), nil
	}

	log.Printf("[Skills] Script completed: %s (%d bytes output)", absScript, len(output))
	return result, nil
}
