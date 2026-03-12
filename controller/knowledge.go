package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/vars"

	"github.com/gin-gonic/gin"
)

type Knowledge struct{}

func NewKnowledgeController() *Knowledge {
	return &Knowledge{}
}

// AddDocument 添加知识库文档
func (k *Knowledge) AddDocument(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.AddKnowledgeDocumentRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if req.Source == "" {
		req.Source = "manual"
	}
	err := vars.KnowledgeService.AddDocument(c.Request.Context(), req.Title, req.Content, req.Source, req.Category)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

// DeleteDocument 删除知识库文档
func (k *Knowledge) DeleteDocument(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.DeleteKnowledgeRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	ctx := c.Request.Context()
	var err error
	if req.ID > 0 {
		err = vars.KnowledgeService.DeleteDocumentByID(ctx, req.ID)
	} else if req.Title != "" {
		err = vars.KnowledgeService.DeleteDocument(ctx, req.Title)
	} else {
		resp.ToErrorResponse(errors.New("请提供 id 或 title"))
		return
	}
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

// ListDocuments 获取知识库文档列表
func (k *Knowledge) ListDocuments(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.ListKnowledgeRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	docs, total, err := vars.KnowledgeService.ListDocuments(c.Request.Context(), req.Category, req.Page, req.PageSize)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponseList(docs, total)
}

// GetCategories 获取知识库分类列表
func (k *Knowledge) GetCategories(c *gin.Context) {
	resp := appx.NewResponse(c)
	categories, err := vars.KnowledgeService.GetCategories(c.Request.Context())
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(categories)
}

// SearchKnowledge 搜索知识库
func (k *Knowledge) SearchKnowledge(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.SearchKnowledgeRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}
	results, err := vars.KnowledgeService.SearchKnowledge(c.Request.Context(), req.Query, req.Category, req.Limit)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(results)
}

// ReindexAll 重建知识库索引
func (k *Knowledge) ReindexAll(c *gin.Context) {
	resp := appx.NewResponse(c)
	go func() {
		if err := vars.KnowledgeService.ReindexAll(c.Request.Context()); err != nil {
			// 异步执行，日志记录即可
		}
	}()
	resp.ToResponse("reindex started")
}

// --- 记忆管理接口 ---

// SaveMemory 手动保存记忆
func (k *Knowledge) SaveMemory(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.SaveMemoryRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if req.Importance <= 0 {
		req.Importance = 5
	}
	memory := &model.Memory{
		ContactWxID: req.ContactWxID,
		ChatRoomID:  req.ChatRoomID,
		Type:        model.MemoryType(req.Type),
		Key:         req.Key,
		Content:     req.Content,
		Importance:  req.Importance,
	}
	err := vars.MemoryService.SaveManualMemory(c.Request.Context(), memory)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(memory)
}

// SearchMemory 搜索记忆
func (k *Knowledge) SearchMemory(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.SearchMemoryRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	memories, err := vars.MemoryService.GetRelevantMemories(c.Request.Context(), req.ContactWxID, req.Query, req.Limit)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(memories)
}

// DeleteMemory 删除记忆
func (k *Knowledge) DeleteMemory(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.DeleteMemoryRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := vars.MemoryService.DeleteMemory(c.Request.Context(), req.ID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
