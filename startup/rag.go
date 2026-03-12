package startup

import (
	"context"
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/qdrantx"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

// InitRAGService 初始化 RAG 相关服务（Qdrant、Embedding、VectorStore、Memory、RAG、Knowledge）
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

	// 初始化向量集合
	if err := qdrantClient.InitCollections(ctx); err != nil {
		return err
	}
	log.Println("Qdrant 连接成功，向量集合已初始化")

	// 2. 获取 AI 配置（BaseURL、APIKey）用于 Embedding
	globalSettings, err := service.NewGlobalSettingsService(ctx).GetGlobalSettings()
	if err != nil {
		return err
	}
	if globalSettings.ChatBaseURL == "" || globalSettings.ChatAPIKey == "" {
		log.Println("[RAG] AI 配置未设置（ChatBaseURL/ChatAPIKey），RAG 服务跳过初始化")
		return nil
	}

	// 3. 初始化 Embedding 服务
	embeddingSvc := service.NewEmbeddingService(globalSettings.ChatBaseURL, globalSettings.ChatAPIKey)

	// 4. 初始化 VectorStore 服务
	vectorStoreSvc := service.NewVectorStoreService(qdrantClient, embeddingSvc)

	// 5. 初始化 Memory 服务
	aiModel := globalSettings.ChatModel
	if aiModel == "" {
		aiModel = "gpt-4o-mini"
	}
	memorySvc := service.NewMemoryService(
		vars.DB, vectorStoreSvc, embeddingSvc,
		globalSettings.ChatBaseURL, globalSettings.ChatAPIKey, aiModel,
	)
	vars.MemoryService = memorySvc

	// 6. 初始化 RAG 服务
	vars.RAGService = service.NewRAGService(vars.DB, memorySvc, vectorStoreSvc)

	// 7. 初始化 Knowledge 服务
	vars.KnowledgeService = service.NewKnowledgeService(vars.DB, vectorStoreSvc)

	// 8. AutoMigrate 新模型
	if err := vars.DB.AutoMigrate(
		&model.Memory{},
		&model.ConversationSession{},
		&model.KnowledgeDocument{},
		&model.EmbeddingTask{},
	); err != nil {
		return err
	}
	log.Println("RAG 服务初始化完成")

	return nil
}
