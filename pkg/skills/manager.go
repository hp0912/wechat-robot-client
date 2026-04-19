package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"wechat-robot-client/pkg/robotctx"

	"github.com/google/shlex"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// SkillRepository 数据库持久化接口（由 repository 层实现）
type SkillRepository interface {
	// FindAll 查询所有 Skill 记录
	FindAll() ([]SkillRecord, error)
	// Upsert 插入或更新一条记录
	Upsert(record SkillRecord) error
	// Delete 按名称删除
	Delete(name string) error
}

// SkillRecord 存储层的 Skill 记录（与 GORM model 解耦）
type SkillRecord struct {
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Enabled     bool        `json:"enabled"`
	Source      SkillSource `json:"source"`
	EnvVars     []EnvVar    `json:"env_vars,omitempty"`
	InstalledAt time.Time   `json:"installed_at"`
}

// SkillsManager Skills 管理器，负责发现、加载、激活、安装技能
type SkillsManager struct {
	// 已加载的所有 Skill（name -> Skill）
	skills map[string]*Skill
	mu     sync.RWMutex

	// Skill 存储根目录
	baseDir string

	// 数据库持久化
	repo SkillRepository

	// Installer
	installer *Installer
}

// ToolNameActivate activate_skill 工具名称
const ToolNameActivate = "activate_skill"

// ToolNameReadResource read_skill_resource 工具名称
const ToolNameReadResource = "read_skill_resource"

// ToolNameExecuteScript execute_skill_script 工具名称
const ToolNameExecuteScript = "execute_skill_script"

// NewSkillsManager 创建 Skills 管理器
func NewSkillsManager(baseDir string, repo SkillRepository) *SkillsManager {
	return &SkillsManager{
		skills:    make(map[string]*Skill),
		baseDir:   baseDir,
		repo:      repo,
		installer: NewInstaller(baseDir),
	}
}

// Initialize 初始化管理器：扫描目录、加载所有 Skill 元数据
func (m *SkillsManager) Initialize() error {
	// 确保目录存在
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	// 从数据库加载配置
	records, err := m.repo.FindAll()
	if err != nil {
		log.Printf("[Skills] Warning: failed to load config from DB: %v", err)
		records = nil
	}

	// 发现磁盘上的 Skill
	paths, err := DiscoverSkills(m.baseDir)
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	// 构建配置索引
	configIndex := make(map[string]*SkillRecord)
	for i := range records {
		configIndex[records[i].Name] = &records[i]
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, path := range paths {
		skill, err := LoadSkillFull(path)
		if err != nil {
			log.Printf("[Skills] Failed to load skill at %s: %v", path, err)
			continue
		}

		// 从配置中恢复状态
		if entry, ok := configIndex[skill.Name]; ok {
			skill.Enabled = entry.Enabled
			skill.Source = entry.Source
			skill.EnvVars = entry.EnvVars
			skill.InstalledAt = entry.InstalledAt
		} else {
			// 新发现的 Skill 默认启用
			skill.Enabled = true
			skill.InstalledAt = time.Now()
			skill.Source = SkillSource{Type: "local"}
		}

		m.skills[skill.Name] = skill
		log.Printf("[Skills] Loaded skill: %s (%s)", skill.Name, skill.Description)
	}

	// 保存最新配置到数据库
	m.syncToDB()

	log.Printf("[Skills] Manager initialized with %d skills", len(m.skills))
	return nil
}

func (m *SkillsManager) Shutdown() error {
	return nil
}

// GetAllSummaries 获取所有已启用 Skill 的摘要（用于注入 system prompt）
func (m *SkillsManager) GetAllSummaries() []SkillSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var summaries []SkillSummary
	for _, skill := range m.skills {
		if !skill.Enabled {
			continue
		}
		summaries = append(summaries, SkillSummary{
			Name:        skill.Name,
			Description: skill.Description,
		})
	}
	return summaries
}

// MatchSkills 根据用户消息匹配可能相关的 Skill 名称列表
// 返回所有已启用 Skill 的名称，由 AI 大模型自行决定激活哪些
func (m *SkillsManager) MatchSkills() []SkillSummary {
	return m.GetAllSummaries()
}

// ActivateSkill 激活 Skill：加载完整的 instructions 返回给调用方
func (m *SkillsManager) ActivateSkill(name string) (*Skill, error) {
	m.mu.RLock()
	skill, ok := m.skills[name]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("skill '%s' not found", name)
	}
	if !skill.Enabled {
		return nil, fmt.Errorf("skill '%s' is disabled", name)
	}

	return skill, nil
}

// ReadResource 读取 Skill 中的附属资源文件
func (m *SkillsManager) ReadResource(skillName, relativePath string) (string, error) {
	m.mu.RLock()
	skill, ok := m.skills[skillName]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("skill '%s' not found", skillName)
	}

	return ReadSkillResource(skill.Path, relativePath)
}

// InstallFromGit 从 Git 仓库安装 Skill
func (m *SkillsManager) InstallFromGit(req SkillInstallRequest) (*Skill, error) {
	if req.Ref == "" {
		req.Ref = "main"
	}

	skillDir, err := m.installer.InstallFromGit(req.RepoURL, req.SubPath, req.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to install skill: %w", err)
	}

	// 加载 Skill
	skill, err := LoadSkillFull(skillDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load installed skill: %w", err)
	}

	skill.Enabled = true
	skill.InstalledAt = time.Now()
	skill.Source = SkillSource{
		Type:    "git",
		RepoURL: req.RepoURL,
		SubPath: req.SubPath,
		Ref:     req.Ref,
	}

	m.mu.Lock()
	m.skills[skill.Name] = skill
	m.mu.Unlock()

	// 保存配置到数据库
	m.saveSkillToDB(skill)

	log.Printf("[Skills] Installed skill: %s from %s", skill.Name, req.RepoURL)
	return skill, nil
}

// Uninstall 卸载 Skill
func (m *SkillsManager) Uninstall(name string) error {
	m.mu.Lock()
	skill, ok := m.skills[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("skill '%s' not found", name)
	}
	delete(m.skills, name)
	m.mu.Unlock()

	// 删除磁盘文件
	if err := os.RemoveAll(skill.Path); err != nil {
		log.Printf("[Skills] Warning: failed to remove skill directory: %v", err)
	}

	// 从数据库删除
	if err := m.repo.Delete(name); err != nil {
		log.Printf("[Skills] Warning: failed to delete skill from DB: %v", err)
	}

	log.Printf("[Skills] Uninstalled skill: %s", name)
	return nil
}

// Enable 启用 Skill
func (m *SkillsManager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	skill, ok := m.skills[name]
	if !ok {
		return fmt.Errorf("skill '%s' not found", name)
	}

	skill.Enabled = true

	m.saveSkillToDB(skill)
	return nil
}

// Disable 禁用 Skill
func (m *SkillsManager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	skill, ok := m.skills[name]
	if !ok {
		return fmt.Errorf("skill '%s' not found", name)
	}

	skill.Enabled = false

	m.saveSkillToDB(skill)
	return nil
}

// GetSkill 获取单个 Skill 信息
func (m *SkillsManager) GetSkill(name string) (*Skill, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	skill, ok := m.skills[name]
	return skill, ok
}

// GetAllSkills 获取所有 Skill
func (m *SkillsManager) GetAllSkills() []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Skill, 0, len(m.skills))
	for _, skill := range m.skills {
		result = append(result, skill)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// GetSkillCount 获取已加载的 Skill 数量
func (m *SkillsManager) GetSkillCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.skills)
}

// BuildSystemPrompt 构建注入到 system prompt 中的 Skills 部分
func (m *SkillsManager) BuildSystemPrompt() string {
	summaries := m.GetAllSummaries()
	if len(summaries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n<available_skills>\n")

	for _, s := range summaries {
		sb.WriteString("  <skill>\n    <name>")
		sb.WriteString(s.Name)
		sb.WriteString("</name>\n    <description>")
		sb.WriteString(s.Description)
		sb.WriteString("</description>\n  </skill>\n")
	}

	sb.WriteString(`</available_skills>

当你判断用户的任务与某个 Skill 相关时，请调用 activate_skill 工具来加载该 Skill 的完整指令。
加载后请严格按照 Skill 指令执行任务。
如果需要读取 Skill 附带的资源文件（如 scripts/、references/ 等），请调用 read_skill_resource 工具。
如果 Skill 指令要求运行脚本（如 Python/Shell 脚本），请调用 execute_skill_script 工具执行。
`)

	return sb.String()
}

// GetOpenAITools 返回 Skills 相关的 OpenAI Tool 定义
func (m *SkillsManager) GetOpenAITools() []openai.Tool {
	summaries := m.GetAllSummaries()
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
func (m *SkillsManager) IsSkillTool(toolName string) bool {
	return toolName == ToolNameActivate || toolName == ToolNameReadResource || toolName == ToolNameExecuteScript
}

// ExecuteToolCall 执行 Skills 工具调用，返回结果字符串
func (m *SkillsManager) ExecuteToolCall(robotCtx robotctx.RobotContext, toolCall openai.ToolCall) (string, error) {
	switch toolCall.Function.Name {
	case ToolNameActivate:
		return m.executeActivate(toolCall.Function.Arguments)
	case ToolNameReadResource:
		return m.executeReadResource(toolCall.Function.Arguments)
	case ToolNameExecuteScript:
		return m.executeScript(robotCtx, toolCall.Function.Arguments)
	default:
		return "", fmt.Errorf("unknown skill tool: %s", toolCall.Function.Name)
	}
}

// executeActivate 执行 activate_skill
func (m *SkillsManager) executeActivate(argsJSON string) (string, error) {
	var args struct {
		SkillName string `json:"skill_name"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse activate_skill args: %w", err)
	}

	skill, err := m.ActivateSkill(args.SkillName)
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
func (m *SkillsManager) executeReadResource(argsJSON string) (string, error) {
	var args struct {
		SkillName string `json:"skill_name"`
		FilePath  string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse read_skill_resource args: %w", err)
	}

	content, err := m.ReadResource(args.SkillName, args.FilePath)
	if err != nil {
		return "", err
	}

	log.Printf("[Skills] Read resource: %s / %s (%d bytes)", args.SkillName, args.FilePath, len(content))
	return content, nil
}

// executeScript 执行 execute_skill_script
func (m *SkillsManager) executeScript(robotCtx robotctx.RobotContext, argsJSON string) (string, error) {
	var args struct {
		SkillName  string `json:"skill_name"`
		ScriptPath string `json:"script_path"`
		Args       string `json:"args"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse execute_skill_script args: %w", err)
	}

	m.mu.RLock()
	skill, ok := m.skills[args.SkillName]
	m.mu.RUnlock()
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
		parsedArgs, err := shlex.Split(args.Args)
		if err != nil {
			return "", fmt.Errorf("failed to parse script args: %w", err)
		}
		cmdArgs = append(cmdArgs, parsedArgs...)
	}

	log.Printf("[Skills] Executing script: %s (args: %s)", absScript, args.Args)

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
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

// syncToDB 将所有内存中的 Skill 同步到数据库
func (m *SkillsManager) syncToDB() {
	for _, skill := range m.skills {
		m.saveSkillToDB(skill)
	}
}

// saveSkillToDB 将单个 Skill 保存到数据库
func (m *SkillsManager) saveSkillToDB(skill *Skill) {
	record := SkillRecord{
		Name:        skill.Name,
		Path:        skill.Path,
		Enabled:     skill.Enabled,
		Source:      skill.Source,
		EnvVars:     skill.EnvVars,
		InstalledAt: skill.InstalledAt,
	}
	if err := m.repo.Upsert(record); err != nil {
		log.Printf("[Skills] Warning: failed to save skill '%s' to DB: %v", skill.Name, err)
	}
}

// SetEnvVars 设置 Skill 的环境变量
func (m *SkillsManager) SetEnvVars(name string, envVars []EnvVar) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	skill, ok := m.skills[name]
	if !ok {
		return fmt.Errorf("skill '%s' not found", name)
	}

	skill.EnvVars = envVars
	m.saveSkillToDB(skill)
	return nil
}

// UpdateSkill 热更新 Skill（从 Git 重新拉取最新版本）
func (m *SkillsManager) UpdateSkill(name string) (*Skill, error) {
	m.mu.RLock()
	existing, ok := m.skills[name]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("skill '%s' not found", name)
	}

	if existing.Source.Type != "git" {
		return nil, fmt.Errorf("skill '%s' is not installed from git, cannot update", name)
	}

	// 重新从 Git 安装（InstallFromGit 内部会先删再装）
	req := SkillInstallRequest{
		RepoURL: existing.Source.RepoURL,
		SubPath: existing.Source.SubPath,
		Ref:     existing.Source.Ref,
	}

	skillDir, err := m.installer.InstallFromGit(req.RepoURL, req.SubPath, req.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to update skill from git: %w", err)
	}

	// 重新加载 Skill
	skill, err := LoadSkillFull(skillDir)
	if err != nil {
		return nil, fmt.Errorf("failed to reload skill after update: %w", err)
	}

	// 保留原有状态
	skill.Enabled = existing.Enabled
	skill.InstalledAt = existing.InstalledAt
	skill.Source = existing.Source
	skill.EnvVars = existing.EnvVars

	m.mu.Lock()
	m.skills[skill.Name] = skill
	m.mu.Unlock()

	// 保存到数据库
	m.saveSkillToDB(skill)

	log.Printf("[Skills] Updated skill: %s from %s", skill.Name, req.RepoURL)
	return skill, nil
}
