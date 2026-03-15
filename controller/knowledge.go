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

// UpdateDocument 更新知识库文档
func (k *Knowledge) UpdateDocument(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.UpdateKnowledgeDocumentRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if req.Source == "" {
		req.Source = "manual"
	}
	err := vars.KnowledgeService.UpdateDocument(c.Request.Context(), req.ID, req.Title, req.Content, req.Source)
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

// --- 图片知识库接口 ---

// AddImageDocument 添加图片知识库文档
func (k *Knowledge) AddImageDocument(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.AddImageKnowledgeRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if vars.ImageKnowledgeService == nil {
		resp.ToErrorResponse(errors.New("图片知识库服务未初始化，请先配置图片嵌入模型"))
		return
	}
	err := vars.ImageKnowledgeService.AddImageDocument(c.Request.Context(), req.Title, req.Description, req.ImageURL, req.Category)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

// DeleteImageDocument 删除图片知识库文档
func (k *Knowledge) DeleteImageDocument(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.DeleteImageKnowledgeRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if vars.ImageKnowledgeService == nil {
		resp.ToErrorResponse(errors.New("图片知识库服务未初始化"))
		return
	}
	ctx := c.Request.Context()
	var err error
	if req.ID > 0 {
		err = vars.ImageKnowledgeService.DeleteImageDocumentByID(ctx, req.ID)
	} else if req.Title != "" {
		err = vars.ImageKnowledgeService.DeleteImageDocument(ctx, req.Title)
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

// ListImageDocuments 获取图片知识库文档列表
func (k *Knowledge) ListImageDocuments(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.ListImageKnowledgeRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if vars.ImageKnowledgeService == nil {
		resp.ToErrorResponse(errors.New("图片知识库服务未初始化"))
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	docs, total, err := vars.ImageKnowledgeService.ListImageDocuments(c.Request.Context(), req.Category, req.Page, req.PageSize)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponseList(docs, total)
}

// SearchImageByText 以文搜图
func (k *Knowledge) SearchImageByText(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.SearchImageKnowledgeByTextRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if vars.ImageKnowledgeService == nil {
		resp.ToErrorResponse(errors.New("图片知识库服务未初始化"))
		return
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}
	results, err := vars.ImageKnowledgeService.SearchByText(c.Request.Context(), req.Query, req.Category, req.Limit)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(results)
}

// SearchImageByImage 以图搜图
func (k *Knowledge) SearchImageByImage(c *gin.Context) {
	resp := appx.NewResponse(c)
	var req dto.SearchImageKnowledgeByImageRequest
	if ok, _ := appx.BindAndValid(c, &req); !ok {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if vars.ImageKnowledgeService == nil {
		resp.ToErrorResponse(errors.New("图片知识库服务未初始化"))
		return
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}
	results, err := vars.ImageKnowledgeService.SearchByImage(c.Request.Context(), req.ImageURL, req.Category, req.Limit)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(results)
}

// ReindexAllImages 重建图片知识库索引
func (k *Knowledge) ReindexAllImages(c *gin.Context) {
	resp := appx.NewResponse(c)
	if vars.ImageKnowledgeService == nil {
		resp.ToErrorResponse(errors.New("图片知识库服务未初始化"))
		return
	}
	go func() {
		if err := vars.ImageKnowledgeService.ReindexAll(c.Request.Context()); err != nil {
			// 异步执行，日志记录即可
		}
	}()
	resp.ToResponse("image reindex started")
}
