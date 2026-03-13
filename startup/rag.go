package startup

import (
	"context"
	"log"
	"wechat-robot-client/pkg/qdrantx"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

func ptrStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func ptrIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

// InitRAGService 初始化 RAG 相关服务（Qdrant、Embedding、VectorStore、Memory、RAG、Knowledge、ImageKnowledge）
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

	// 2. 获取 AI 配置
	globalSettings, err := service.NewGlobalSettingsService(ctx).GetGlobalSettings()
	if err != nil {
		return err
	}
	textEmbeddingModel := ""
	imageEmbeddingModel := ""
	imageEmbeddingBaseURL := ""
	imageEmbeddingAPIKey := ""
	imageEmbeddingDimension := 0
	if globalSettings != nil {
		textEmbeddingModel = ptrStringValue(globalSettings.TextEmbeddingModel)
		imageEmbeddingModel = ptrStringValue(globalSettings.ImageEmbeddingModel)
		imageEmbeddingBaseURL = ptrStringValue(globalSettings.ImageEmbeddingBaseURL)
		imageEmbeddingAPIKey = ptrStringValue(globalSettings.ImageEmbeddingAPIKey)
		imageEmbeddingDimension = ptrIntValue(globalSettings.ImageEmbeddingDimension)
	}

	// 3. 初始化文本向量集合
	textDimension := uint64(qdrantx.DefaultEmbeddingDimension)
	if err := qdrantClient.InitCollections(ctx, textDimension); err != nil {
		return err
	}
	log.Println("Qdrant 连接成功，文本向量集合已初始化")

	// 4. 初始化图片向量集合（如果配置了图片嵌入模型）
	if globalSettings != nil && imageEmbeddingModel != "" && imageEmbeddingDimension > 0 {
		if err := qdrantClient.InitCollection(ctx, qdrantx.CollectionImageKnowledge, uint64(imageEmbeddingDimension)); err != nil {
			return err
		}
		log.Println("Qdrant 图片向量集合已初始化")
	}

	if globalSettings == nil || globalSettings.ChatBaseURL == "" || globalSettings.ChatAPIKey == "" {
		log.Println("[RAG] AI 配置未设置（ChatBaseURL/ChatAPIKey），RAG 服务跳过初始化")
		return nil
	}

	// 5. 初始化文本 Embedding 服务（支持可配置模型）
	embeddingSvc := service.NewEmbeddingService(globalSettings.ChatBaseURL, globalSettings.ChatAPIKey, textEmbeddingModel)

	// 6. 初始化 VectorStore 服务
	vectorStoreSvc := service.NewVectorStoreService(qdrantClient, embeddingSvc)

	// 7. 初始化图片 Embedding 服务（如果配置了图片嵌入模型）
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
	}

	// 8. 初始化 Memory 服务
	aiModel := globalSettings.ChatModel
	if aiModel == "" {
		aiModel = "gpt-4o-mini"
	}
	memorySvc := service.NewMemoryService(
		vars.DB, vectorStoreSvc, embeddingSvc,
		globalSettings.ChatBaseURL, globalSettings.ChatAPIKey, aiModel,
	)
	vars.MemoryService = memorySvc

	// 9. 初始化 RAG 服务
	vars.RAGService = service.NewRAGService(vars.DB, memorySvc, vectorStoreSvc)

	// 10. 初始化 Knowledge 服务
	vars.KnowledgeService = service.NewKnowledgeService(vars.DB, vectorStoreSvc)
	log.Println("RAG 服务初始化完成")

	return nil
}
