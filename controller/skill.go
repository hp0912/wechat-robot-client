package controller

import (
	"errors"

	"github.com/gin-gonic/gin"

	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/pkg/skills"
	"wechat-robot-client/vars"
)

type SkillController struct{}

func NewSkillController() *SkillController {
	return &SkillController{}
}

// GetAllSkills 获取所有已安装的 Skills
func (s *SkillController) GetAllSkills(c *gin.Context) {
	resp := appx.NewResponse(c)

	if vars.SkillService == nil {
		resp.ToErrorResponse(errors.New("Skills 服务未初始化"))
		return
	}

	allSkills := vars.SkillService.GetAllSkills()
	resp.ToResponse(allSkills)
}

// GetSkill 获取单个 Skill 详情
func (s *SkillController) GetSkill(c *gin.Context) {
	var req struct {
		Name string `form:"name" json:"name" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	if vars.SkillService == nil {
		resp.ToErrorResponse(errors.New("Skills 服务未初始化"))
		return
	}

	skill, ok := vars.SkillService.GetSkill(req.Name)
	if !ok {
		resp.ToErrorResponse(errors.New("Skill 不存在"))
		return
	}

	resp.ToResponse(skill)
}

// InstallSkill 从 Git 仓库安装 Skill
func (s *SkillController) InstallSkill(c *gin.Context) {
	var req struct {
		// GitHub URL（完整路径，如 https://github.com/anthropics/skills/tree/main/skills/pptx）
		URL string `json:"url"`
		// 或者分别指定
		RepoURL string `json:"repo_url"`
		SubPath string `json:"sub_path"`
		Ref     string `json:"ref"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	if vars.SkillService == nil {
		resp.ToErrorResponse(errors.New("Skills 服务未初始化"))
		return
	}

	var (
		skill *skills.Skill
		err   error
	)

	if req.URL != "" {
		skill, err = vars.SkillService.InstallSkillFromURL(req.URL)
	} else if req.RepoURL != "" && req.SubPath != "" {
		skill, err = vars.SkillService.InstallSkill(skills.SkillInstallRequest{
			RepoURL: req.RepoURL,
			SubPath: req.SubPath,
			Ref:     req.Ref,
		})
	} else {
		resp.ToErrorResponse(errors.New("请提供 url 或 repo_url + sub_path"))
		return
	}

	if err != nil {
		resp.ToErrorResponse(err)
		return
	}

	resp.ToResponse(skill)
}

// UninstallSkill 卸载 Skill
func (s *SkillController) UninstallSkill(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	if vars.SkillService == nil {
		resp.ToErrorResponse(errors.New("Skills 服务未初始化"))
		return
	}

	if err := vars.SkillService.UninstallSkill(req.Name); err != nil {
		resp.ToErrorResponse(err)
		return
	}

	resp.ToResponse(nil)
}

// EnableSkill 启用 Skill
func (s *SkillController) EnableSkill(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	if vars.SkillService == nil {
		resp.ToErrorResponse(errors.New("Skills 服务未初始化"))
		return
	}

	if err := vars.SkillService.EnableSkill(req.Name); err != nil {
		resp.ToErrorResponse(err)
		return
	}

	resp.ToResponse(nil)
}

// DisableSkill 禁用 Skill
func (s *SkillController) DisableSkill(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	if vars.SkillService == nil {
		resp.ToErrorResponse(errors.New("Skills 服务未初始化"))
		return
	}

	if err := vars.SkillService.DisableSkill(req.Name); err != nil {
		resp.ToErrorResponse(err)
		return
	}

	resp.ToResponse(nil)
}

// UpdateSkill 热更新 Skill（从 Git 重新拉取最新版本）
func (s *SkillController) UpdateSkill(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	if vars.SkillService == nil {
		resp.ToErrorResponse(errors.New("Skills 服务未初始化"))
		return
	}

	skill, err := vars.SkillService.UpdateSkill(req.Name)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}

	resp.ToResponse(skill)
}
