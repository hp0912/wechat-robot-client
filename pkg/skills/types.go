package skills

import "time"

// EnvVar Skill 的单个环境变量
type EnvVar struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// SkillMetadata SKILL.md frontmatter 解析后的元数据
type SkillMetadata struct {
	Name          string            `yaml:"name" json:"name"`
	Description   string            `yaml:"description" json:"description"`
	License       string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty" json:"allowed_tools,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// Skill 一个已加载的完整 Skill
type Skill struct {
	// 元数据（来自 SKILL.md frontmatter）
	SkillMetadata `json:"metadata"`

	// 完整 SKILL.md body（markdown 指令部分）
	Instructions string `json:"instructions"`

	// Skill 在磁盘上的绝对路径
	Path string `json:"path"`

	// 来源信息
	Source SkillSource `json:"source"`

	// 安装时间
	InstalledAt time.Time `json:"installed_at"`

	// 是否已启用
	Enabled bool `json:"enabled"`

	// 环境变量列表
	EnvVars []EnvVar `json:"env_vars,omitempty"`
}

// SkillSource 技能来源
type SkillSource struct {
	// 类型：local / git
	Type string `json:"type"`
	// Git 仓库 URL（如果通过 Git 安装）
	RepoURL string `json:"repo_url,omitempty"`
	// 仓库中的子路径
	SubPath string `json:"sub_path,omitempty"`
	// Git ref（branch/tag/commit）
	Ref string `json:"ref,omitempty"`
}

// SkillSummary 轻量摘要，用于注入 system prompt
type SkillSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SkillInstallRequest 安装 Skill 的请求
type SkillInstallRequest struct {
	// Git 仓库 URL，例如 https://github.com/anthropics/skills
	RepoURL string `json:"repo_url"`
	// 仓库中的子路径，例如 skills/pptx
	SubPath string `json:"sub_path"`
	// Git 分支/标签，默认 main
	Ref string `json:"ref"`
}
