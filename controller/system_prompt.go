package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type SystemPrompt struct{}

func NewSystemPromptController() *SystemPrompt {
	return &SystemPrompt{}
}

func (p *SystemPrompt) svc(c *gin.Context) *service.SystemPromptService {
	return service.NewSystemPromptService(c.Request.Context())
}

// Create 创建系统提示词
func (p *SystemPrompt) Create(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.CreateSystemPromptRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	prompt, err := p.svc(c).Create(req.Title, req.Content)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(prompt)
}

// Update 更新系统提示词
func (p *SystemPrompt) Update(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.UpdateSystemPromptRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := p.svc(c).Update(req.ID, req.Title, req.Content)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

// Delete 删除系统提示词
func (p *SystemPrompt) Delete(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.DeleteSystemPromptRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := p.svc(c).Delete(req.ID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

// Get 获取系统提示词
func (p *SystemPrompt) Get(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.GetSystemPromptRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	prompt, err := p.svc(c).GetByID(req.ID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(prompt)
}

// List 获取系统提示词列表
func (p *SystemPrompt) List(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.ListSystemPromptRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	prompts, err := p.svc(c).List(req.Keyword)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(prompts)
}
