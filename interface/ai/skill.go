package ai

import (
	"wechat-robot-client/pkg/skills"
)

// SkillService Skills 技能管理服务接口
type SkillService interface {
	// Initialize 初始化 Skills 服务
	Initialize() error
	// GetManager 获取 Skills Manager
	GetManager() *skills.Manager
	// GetExecutor 获取 Skills 工具执行器
	GetExecutor() *skills.SkillToolExecutor
	// InstallSkill 从 Git 仓库安装 Skill
	InstallSkill(req skills.SkillInstallRequest) (*skills.Skill, error)
	// InstallSkillFromURL 从 GitHub URL 安装 Skill
	InstallSkillFromURL(url string) (*skills.Skill, error)
	// UninstallSkill 卸载 Skill
	UninstallSkill(name string) error
	// EnableSkill 启用 Skill
	EnableSkill(name string) error
	// DisableSkill 禁用 Skill
	DisableSkill(name string) error
	// GetAllSkills 获取所有已安装的 Skills
	GetAllSkills() []*skills.Skill
	// GetSkill 获取单个 Skill
	GetSkill(name string) (*skills.Skill, bool)
	// UpdateSkill 热更新 Skill（从 Git 重新拉取）
	UpdateSkill(name string) (*skills.Skill, error)
	// SetEnvVars 设置 Skill 的环境变量列表
	SetEnvVars(name string, envVars []skills.EnvVar) error
}
