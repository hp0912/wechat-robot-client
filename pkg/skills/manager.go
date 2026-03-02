package skills

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager Skills 管理器，负责发现、加载、激活、安装技能
type Manager struct {
	// 已加载的所有 Skill（name -> Skill）
	skills map[string]*Skill
	mu     sync.RWMutex

	// Skill 存储根目录
	baseDir string

	// 配置文件路径
	configPath string

	// Installer
	installer *Installer
}

// NewManager 创建 Skills 管理器
func NewManager(baseDir string) *Manager {
	configPath := filepath.Join(baseDir, "skills.json")
	return &Manager{
		skills:     make(map[string]*Skill),
		baseDir:    baseDir,
		configPath: configPath,
		installer:  NewInstaller(baseDir),
	}
}

// Initialize 初始化管理器：扫描目录、加载所有 Skill 元数据
func (m *Manager) Initialize() error {
	// 确保目录存在
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	// 加载配置
	config, err := m.loadConfig()
	if err != nil {
		log.Printf("[Skills] Warning: failed to load config, will discover from disk: %v", err)
		config = &SkillConfig{}
	}

	// 发现磁盘上的 Skill
	paths, err := DiscoverSkills(m.baseDir)
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	// 构建配置索引
	configIndex := make(map[string]*SkillEntry)
	for i := range config.Skills {
		configIndex[config.Skills[i].Name] = &config.Skills[i]
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

	// 保存最新配置
	if err := m.saveConfig(); err != nil {
		log.Printf("[Skills] Warning: failed to save config: %v", err)
	}

	log.Printf("[Skills] Manager initialized with %d skills", len(m.skills))
	return nil
}

// GetAllSummaries 获取所有已启用 Skill 的摘要（用于注入 system prompt）
func (m *Manager) GetAllSummaries() []SkillSummary {
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
func (m *Manager) MatchSkills() []SkillSummary {
	return m.GetAllSummaries()
}

// ActivateSkill 激活 Skill：加载完整的 instructions 返回给调用方
func (m *Manager) ActivateSkill(name string) (*Skill, error) {
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
func (m *Manager) ReadResource(skillName, relativePath string) (string, error) {
	m.mu.RLock()
	skill, ok := m.skills[skillName]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("skill '%s' not found", skillName)
	}

	return ReadSkillResource(skill.Path, relativePath)
}

// InstallFromGit 从 Git 仓库安装 Skill
func (m *Manager) InstallFromGit(req SkillInstallRequest) (*Skill, error) {
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

	// 保存配置
	if err := m.saveConfig(); err != nil {
		log.Printf("[Skills] Warning: failed to save config after install: %v", err)
	}

	log.Printf("[Skills] Installed skill: %s from %s", skill.Name, req.RepoURL)
	return skill, nil
}

// Uninstall 卸载 Skill
func (m *Manager) Uninstall(name string) error {
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

	// 保存配置
	if err := m.saveConfig(); err != nil {
		log.Printf("[Skills] Warning: failed to save config after uninstall: %v", err)
	}

	log.Printf("[Skills] Uninstalled skill: %s", name)
	return nil
}

// Enable 启用 Skill
func (m *Manager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	skill, ok := m.skills[name]
	if !ok {
		return fmt.Errorf("skill '%s' not found", name)
	}

	skill.Enabled = true

	if err := m.saveConfig(); err != nil {
		log.Printf("[Skills] Warning: failed to save config: %v", err)
	}
	return nil
}

// Disable 禁用 Skill
func (m *Manager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	skill, ok := m.skills[name]
	if !ok {
		return fmt.Errorf("skill '%s' not found", name)
	}

	skill.Enabled = false

	if err := m.saveConfig(); err != nil {
		log.Printf("[Skills] Warning: failed to save config: %v", err)
	}
	return nil
}

// GetSkill 获取单个 Skill 信息
func (m *Manager) GetSkill(name string) (*Skill, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	skill, ok := m.skills[name]
	return skill, ok
}

// GetAllSkills 获取所有 Skill
func (m *Manager) GetAllSkills() []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Skill, 0, len(m.skills))
	for _, skill := range m.skills {
		result = append(result, skill)
	}
	return result
}

// GetSkillCount 获取已加载的 Skill 数量
func (m *Manager) GetSkillCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.skills)
}

// BuildSystemPromptSkillsSection 构建注入到 system prompt 中的 Skills 部分
func (m *Manager) BuildSystemPromptSkillsSection() string {
	summaries := m.GetAllSummaries()
	if len(summaries) == 0 {
		return ""
	}

	var sb []byte
	sb = append(sb, "\n\n<available_skills>\n"...)

	for _, s := range summaries {
		sb = append(sb, fmt.Sprintf(`  <skill>
    <name>%s</name>
    <description>%s</description>
  </skill>
`, s.Name, s.Description)...)
	}

	sb = append(sb, `</available_skills>

当你判断用户的任务与某个 Skill 相关时，请调用 activate_skill 工具来加载该 Skill 的完整指令。
加载后请严格按照 Skill 指令执行任务。
如果需要读取 Skill 附带的资源文件（如 scripts/、references/ 等），请调用 read_skill_resource 工具。
`...)

	return string(sb)
}

// BuildSkillTools 构建 Skills 相关的 OpenAI Function Tools 定义
// 返回两个工具：activate_skill 和 read_skill_resource
func (m *Manager) BuildSkillTools() []map[string]interface{} {
	summaries := m.GetAllSummaries()
	if len(summaries) == 0 {
		return nil
	}

	// 构建可用 Skill 名称列表用于枚举
	var skillNames []string
	for _, s := range summaries {
		skillNames = append(skillNames, s.Name)
	}

	tools := []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "activate_skill",
				"description": "激活一个 Agent Skill，加载其完整的操作指令到上下文中。当你判断用户任务与某个可用 Skill 相关时调用此工具。",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"skill_name": map[string]interface{}{
							"type":        "string",
							"description": "要激活的 Skill 名称",
							"enum":        skillNames,
						},
					},
					"required": []string{"skill_name"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "read_skill_resource",
				"description": "读取已激活 Skill 中的附属资源文件（如 scripts/、references/、assets/ 下的文件）。当 Skill 指令中引用了外部文件时调用此工具。",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"skill_name": map[string]interface{}{
							"type":        "string",
							"description": "Skill 名称",
						},
						"file_path": map[string]interface{}{
							"type":        "string",
							"description": "要读取的文件相对路径，例如 scripts/extract.py 或 references/REFERENCE.md",
						},
					},
					"required": []string{"skill_name", "file_path"},
				},
			},
		},
	}

	return tools
}

// loadConfig 加载配置文件
func (m *Manager) loadConfig() (*SkillConfig, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &SkillConfig{}, nil
		}
		return nil, err
	}

	var config SkillConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// saveConfig 保存配置文件（调用方须持有锁或在锁内调用）
func (m *Manager) saveConfig() error {
	config := SkillConfig{
		Skills: make([]SkillEntry, 0, len(m.skills)),
	}

	for _, skill := range m.skills {
		config.Skills = append(config.Skills, SkillEntry{
			Name:        skill.Name,
			Path:        skill.Path,
			Enabled:     skill.Enabled,
			Source:      skill.Source,
			InstalledAt: skill.InstalledAt,
		})
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}
