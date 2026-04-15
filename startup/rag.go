package startup

import (
	"context"
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/qdrantx"
	"wechat-robot-client/service"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"
)

// InitRAGService 初始化 RAG 相关服务
func InitRAGService() error {
	ctx := context.Background()

	// 1. 初始化 Qdrant 客户端
	qdrantClient, err := qdrantx.NewQdrantClient(
		vars.QdrantSettings.Host,
		vars.QdrantSettings.Port,
		vars.QdrantSettings.ApiKey,
	)
	if err != nil {
		return err
	}
	vars.QdrantClient = qdrantClient

	// 2. 获取全局配置，用于确定向量维度并初始化依赖配置的服务
	globalSettings, err := service.NewGlobalSettingsService(ctx).GetGlobalSettings()
	if err != nil {
		return err
	}

	// 3. 初始化文本向量集合
	textEmbeddingDim := uint64(2048)
	if globalSettings != nil && globalSettings.TextEmbeddingDimension != nil && *globalSettings.TextEmbeddingDimension > 0 {
		textEmbeddingDim = uint64(*globalSettings.TextEmbeddingDimension)
	}
	if err := qdrantClient.InitCollections(ctx, textEmbeddingDim); err != nil {
		return err
	}
	log.Println("Qdrant 连接成功，文本向量集合已初始化")
	if err := reloadRAGServices(globalSettings); err != nil {
		return err
	}

	// 4. 注册全局配置变更回调，配置修改后自动重新初始化 RAG 服务
	vars.SettingsObserver.Register("RAG服务", func(settings *model.GlobalSettings) error {
		return reloadRAGServices(settings)
	})

	return nil
}

// reloadRAGServices 根据全局配置（重新）初始化 RAG 相关服务
// 在启动时和全局配置变更时均会调用
func reloadRAGServices(globalSettings *model.GlobalSettings) error {
	ctx := context.Background()

	if vars.QdrantClient == nil {
		log.Println("[RAG] Qdrant 客户端未初始化，跳过 RAG 服务重载")
		return nil
	}

	textEmbeddingModel := ""
	imageEmbeddingModel := ""
	imageEmbeddingBaseURL := ""
	imageEmbeddingAPIKey := ""
	imageEmbeddingDimension := 0
	textEmbeddingDimension := 2048
	if globalSettings != nil {
		textEmbeddingModel = utils.PtrStringValue(globalSettings.TextEmbeddingModel)
		imageEmbeddingModel = utils.PtrStringValue(globalSettings.ImageEmbeddingModel)
		imageEmbeddingBaseURL = utils.PtrStringValue(globalSettings.ImageEmbeddingBaseURL)
		imageEmbeddingAPIKey = utils.PtrStringValue(globalSettings.ImageEmbeddingAPIKey)
		imageEmbeddingDimension = utils.PtrIntValue(globalSettings.ImageEmbeddingDimension)
		if v := utils.PtrIntValue(globalSettings.TextEmbeddingDimension); v > 0 {
			textEmbeddingDimension = v
		}
	}

	// 初始化图片向量集合
	if globalSettings != nil && imageEmbeddingModel != "" && imageEmbeddingDimension > 0 {
		if err := vars.QdrantClient.InitCollection(ctx, qdrantx.CollectionImageKnowledge, uint64(imageEmbeddingDimension)); err != nil {
			return err
		}
		log.Println("Qdrant 图片向量集合已初始化")
	}

	if globalSettings == nil || globalSettings.ChatBaseURL == "" || globalSettings.ChatAPIKey == "" || textEmbeddingModel == "" {
		log.Println("[RAG] AI 配置未设置（ChatBaseURL/ChatAPIKey/TextEmbeddingModel），RAG 服务跳过初始化")
		vars.MemoryService = nil
		vars.RAGService = nil
		vars.KnowledgeService = nil
		vars.ImageKnowledgeService = nil
		return nil
	}

	// 初始化文本 Embedding 服务（支持可配置模型）
	embeddingSvc := service.NewEmbeddingService(globalSettings.ChatBaseURL, globalSettings.ChatAPIKey, textEmbeddingModel, textEmbeddingDimension)

	// 初始化 VectorStore 服务
	vectorStoreSvc := service.NewVectorStoreService(vars.QdrantClient, embeddingSvc)

	// 初始化图片 Embedding 服务（如果配置了图片嵌入模型）
	if imageEmbeddingModel != "" && imageEmbeddingDimension > 0 {
		imageBaseURL := imageEmbeddingBaseURL
		if imageBaseURL == "" {
			imageBaseURL = globalSettings.ChatBaseURL
		}
		imageAPIKey := imageEmbeddingAPIKey
		if imageAPIKey == "" {
			imageAPIKey = globalSettings.ChatAPIKey
		}
		imageEmbeddingSvc := service.NewImageEmbeddingService(imageBaseURL, imageAPIKey, imageEmbeddingModel)
		vectorStoreSvc.SetImageEmbedding(imageEmbeddingSvc)

		// 初始化图片知识库服务
		vars.ImageKnowledgeService = service.NewImageKnowledgeService(vars.DB, vectorStoreSvc)
		log.Println("图片知识库服务初始化完成")
	} else {
		vars.ImageKnowledgeService = nil
	}

	// 初始化 Memory 服务（仅在启用长期记忆时）
	var memorySvc *service.MemoryService
	if globalSettings.MemoryEnabled == nil || *globalSettings.MemoryEnabled {
		aiModel := globalSettings.ChatModel
		if aiModel == "" {
			aiModel = "gpt-4o-mini"
		}
		memorySvc = service.NewMemoryService(
			vars.DB, vectorStoreSvc,
			globalSettings.ChatBaseURL, globalSettings.ChatAPIKey, aiModel,
		)
		log.Println("长期记忆服务已启用")
	} else {
		log.Println("长期记忆服务已禁用")
	}
	if memorySvc != nil {
		vars.MemoryService = memorySvc
	} else {
		vars.MemoryService = nil
	}

	// 初始化 RAG 服务
	vars.RAGService = service.NewRAGService(vars.DB, memorySvc, vectorStoreSvc)

	// 初始化 Knowledge 服务
	vars.KnowledgeService = service.NewKnowledgeService(vars.DB, vectorStoreSvc)
	log.Println("RAG 服务初始化完成")

	return nil
}
