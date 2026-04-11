package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/skills"
	"wechat-robot-client/repository"
	"wechat-robot-client/utils"
)

// SkillService Skills 技能管理服务
type SkillService struct {
	manager *skills.Manager
}

// 确保实现接口
var _ ai.SkillService = (*SkillService)(nil)

// NewSkillService 创建 Skills 服务
func NewSkillService(skillsDir string, db *gorm.DB) *SkillService {
	repo := newSkillRepoAdapter(db)
	manager := skills.NewManager(skillsDir, repo)
	return &SkillService{
		manager: manager,
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

// UpdateSkill 热更新 Skill（从 Git 重新拉取最新版本）
func (s *SkillService) UpdateSkill(name string) (*skills.Skill, error) {
	return s.manager.UpdateSkill(name)
}

// SetEnvVars 设置 Skill 的环境变量
func (s *SkillService) SetEnvVars(name string, envVars []skills.EnvVar) error {
	return s.manager.SetEnvVars(name, envVars)
}

// skillRepoAdapter 将 repository.SkillRepo 适配为 skills.SkillRepository 接口
type skillRepoAdapter struct {
	db *gorm.DB
}

func newSkillRepoAdapter(db *gorm.DB) *skillRepoAdapter {
	return &skillRepoAdapter{db: db}
}

func (a *skillRepoAdapter) FindAll() ([]skills.SkillRecord, error) {
	repo := repository.NewSkillRepo(context.Background(), a.db)
	models, err := repo.FindAll()
	if err != nil {
		return nil, err
	}
	records := make([]skills.SkillRecord, 0, len(models))
	for _, m := range models {
		records = append(records, skills.SkillRecord{
			Name:        m.Name,
			Path:        m.Path,
			Enabled:     m.IsEnabled(),
			Source:      repository.ToSkillSource(m),
			EnvVars:     repository.ToSkillEnvVars(m),
			InstalledAt: utils.PtrTimeValue(m.InstalledAt),
		})
	}
	return records, nil
}

func (a *skillRepoAdapter) Upsert(record skills.SkillRecord) error {
	repo := repository.NewSkillRepo(context.Background(), a.db)
	enabled := record.Enabled
	installedAt := record.InstalledAt
	now := time.Now()
	m := &model.Skill{
		Name:        record.Name,
		Path:        record.Path,
		Enabled:     &enabled,
		SourceType:  model.SkillSourceType(record.Source.Type),
		Source:      repository.SourceToJSON(record.Source),
		EnvVars:     repository.EnvVarsToJSON(record.EnvVars),
		InstalledAt: &installedAt,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	existing, err := repo.FindByName(record.Name)
	if err != nil {
		return err
	}
	if existing != nil {
		m.ID = existing.ID
		m.CreatedAt = existing.CreatedAt
		return repo.Update(m)
	}
	return repo.Create(m)
}

func (a *skillRepoAdapter) Delete(name string) error {
	repo := repository.NewSkillRepo(context.Background(), a.db)
	return repo.Delete(name)
}
