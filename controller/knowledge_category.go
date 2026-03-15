package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"

	"github.com/gin-gonic/gin"
)

type KnowledgeCategory struct{}

func NewKnowledgeCategoryController() *KnowledgeCategory {
	return &KnowledgeCategory{}
}

func (k *KnowledgeCategory) svc() *service.KnowledgeCategoryService {
	return service.NewKnowledgeCategoryService(vars.DB)
}

// Create 创建知识库分类
func (k *KnowledgeCategory) Create(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.CreateKnowledgeCategoryRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	category, err := k.svc().Create(c.Request.Context(), req.Code, req.Name, req.Description)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(category)
}

// Update 更新知识库分类
func (k *KnowledgeCategory) Update(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.UpdateKnowledgeCategoryRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := k.svc().Update(c.Request.Context(), req.ID, req.Name, req.Description)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

// Delete 删除知识库分类
func (k *KnowledgeCategory) Delete(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.DeleteKnowledgeCategoryRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := k.svc().Delete(c.Request.Context(), req.ID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

// List 获取知识库分类列表
func (k *KnowledgeCategory) List(c *gin.Context) {
	resp := appx.NewResponse(c)
	categories, err := k.svc().List(c.Request.Context())
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(categories)
}
