package service

import (
	"fmt"
	"log"

	"wechat-robot-client/interface/ai"
	"wechat-robot-client/pkg/skills"
)

// SkillService Skills 技能管理服务
type SkillService struct {
	manager  *skills.Manager
	executor *skills.SkillToolExecutor
}

// 确保实现接口
var _ ai.SkillService = (*SkillService)(nil)

// NewSkillService 创建 Skills 服务
func NewSkillService(skillsDir string) *SkillService {
	manager := skills.NewManager(skillsDir)
	executor := skills.NewSkillToolExecutor(manager)
	return &SkillService{
		manager:  manager,
		executor: executor,
	}
}

// Initialize 初始化 Skills 服务
func (s *SkillService) Initialize() error {
	log.Println("Initializing Skills Service...")
	return s.manager.Initialize()
}

// GetManager 获取 Manager
func (s *SkillService) GetManager() *skills.Manager {
	return s.manager
}

// GetExecutor 获取 SkillToolExecutor
func (s *SkillService) GetExecutor() *skills.SkillToolExecutor {
	return s.executor
}

// InstallSkill 从 Git 仓库安装 Skill
func (s *SkillService) InstallSkill(req skills.SkillInstallRequest) (*skills.Skill, error) {
	return s.manager.InstallFromGit(req)
}

// InstallSkillFromURL 从 GitHub URL 安装 Skill（自动解析 URL）
func (s *SkillService) InstallSkillFromURL(url string) (*skills.Skill, error) {
	repoURL, subPath, ref := skills.ExtractRepoAndSubPath(url)
	if subPath == "" {
		return nil, fmt.Errorf("cannot determine skill subPath from URL: %s, please provide subPath explicitly", url)
	}
	return s.manager.InstallFromGit(skills.SkillInstallRequest{
		RepoURL: repoURL,
		SubPath: subPath,
		Ref:     ref,
	})
}

// UninstallSkill 卸载 Skill
func (s *SkillService) UninstallSkill(name string) error {
	return s.manager.Uninstall(name)
}

// EnableSkill 启用 Skill
func (s *SkillService) EnableSkill(name string) error {
	return s.manager.Enable(name)
}

// DisableSkill 禁用 Skill
func (s *SkillService) DisableSkill(name string) error {
	return s.manager.Disable(name)
}

// GetAllSkills 获取所有已安装的 Skills
func (s *SkillService) GetAllSkills() []*skills.Skill {
	return s.manager.GetAllSkills()
}

// GetSkill 获取单个 Skill
func (s *SkillService) GetSkill(name string) (*skills.Skill, bool) {
	return s.manager.GetSkill(name)
}
