package service

import (
	"context"
	"fmt"
	"strconv"

	"wechat-robot-client/interface/ai"
	"wechat-robot-client/pkg/qdrantx"

	pb "github.com/qdrant/go-client/qdrant"
)

// VectorStoreService 向量存储服务
type VectorStoreService struct {
	qdrant         *qdrantx.QdrantClient
	embedding      *EmbeddingService
	imageEmbedding *ImageEmbeddingService
}

// NewVectorStoreService 创建向量存储服务
func NewVectorStoreService(qdrant *qdrantx.QdrantClient, embedding *EmbeddingService) *VectorStoreService {
	return &VectorStoreService{
		qdrant:    qdrant,
		embedding: embedding,
	}
}

// SetImageEmbedding 设置图片嵌入服务
func (s *VectorStoreService) SetImageEmbedding(svc *ImageEmbeddingService) {
	s.imageEmbedding = svc
}

// IndexMessage 将消息内容向量化并存入 Qdrant
func (s *VectorStoreService) IndexMessage(ctx context.Context, robotCode string, msgID int64, content, contactWxID, chatRoomID, senderWxID string, createdAt int64) (string, error) {
	vector, err := s.embedding.Embed(ctx, content)
	if err != nil {
		return "", fmt.Errorf("embed message: %w", err)
	}

	id := s.qdrant.GenerateID()
	payload := map[string]*pb.Value{
		"robot_code":            qdrantx.NewPayloadValue(robotCode),
		"msg_id":                qdrantx.NewPayloadIntValue(msgID),
		"content":               qdrantx.NewPayloadValue(content),
		"contact_wxid":          qdrantx.NewPayloadValue(contactWxID),
		"chat_room_id":          qdrantx.NewPayloadValue(chatRoomID),
		"chat_room_member_wxid": qdrantx.NewPayloadValue(senderWxID),
		"created_at":            qdrantx.NewPayloadIntValue(createdAt),
	}

	if err := s.qdrant.Upsert(ctx, qdrantx.CollectionMessages, id, vector, payload); err != nil {
		return "", fmt.Errorf("upsert message vector: %w", err)
	}
	return id, nil
}

// IndexMemory 将记忆内容向量化并存入 Qdrant
func (s *VectorStoreService) IndexMemory(ctx context.Context, robotCode string, memoryID int64, content, wxID, category, chatRoomID string) (string, error) {
	vector, err := s.embedding.Embed(ctx, content)
	if err != nil {
		return "", fmt.Errorf("embed memory: %w", err)
	}

	id := s.qdrant.GenerateID()
	payload := map[string]*pb.Value{
		"robot_code":   qdrantx.NewPayloadValue(robotCode),
		"memory_id":    qdrantx.NewPayloadIntValue(memoryID),
		"category":     qdrantx.NewPayloadValue(category),
		"content":      qdrantx.NewPayloadValue(content),
		"contact_wxid": qdrantx.NewPayloadValue(wxID),
		"chat_room_id": qdrantx.NewPayloadValue(chatRoomID),
	}

	if err := s.qdrant.Upsert(ctx, qdrantx.CollectionMemories, id, vector, payload); err != nil {
		return "", fmt.Errorf("upsert memory vector: %w", err)
	}
	return id, nil
}

// IndexKnowledge 将知识库内容向量化并存入 Qdrant
func (s *VectorStoreService) IndexKnowledge(ctx context.Context, robotCode string, docID int64, category, title, content string) (string, error) {
	vector, err := s.embedding.Embed(ctx, content)
	if err != nil {
		return "", fmt.Errorf("embed knowledge: %w", err)
	}

	id := s.qdrant.GenerateID()
	payload := map[string]*pb.Value{
		"robot_code": qdrantx.NewPayloadValue(robotCode),
		"doc_id":     qdrantx.NewPayloadIntValue(docID),
		"category":   qdrantx.NewPayloadValue(category),
		"title":      qdrantx.NewPayloadValue(title),
		"content":    qdrantx.NewPayloadValue(content),
	}

	if err := s.qdrant.Upsert(ctx, qdrantx.CollectionKnowledge, id, vector, payload); err != nil {
		return "", fmt.Errorf("upsert knowledge vector: %w", err)
	}
	return id, nil
}

// SearchMessages 语义搜索历史消息
func (s *VectorStoreService) SearchMessages(ctx context.Context, robotCode string, query string, contactWxID, chatRoomID string, topK int) ([]ai.VectorSearchResult, error) {
	vector, err := s.embedding.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	var conditions []*pb.Condition
	if robotCode != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("robot_code", robotCode))
	}
	if contactWxID != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("contact_wxid", contactWxID))
	}
	if chatRoomID != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("chat_room_id", chatRoomID))
	}

	var filter *pb.Filter
	if len(conditions) > 0 {
		filter = &pb.Filter{Must: conditions}
	}

	results, err := s.qdrant.Search(ctx, qdrantx.CollectionMessages, vector, uint64(topK), filter)
	if err != nil {
		return nil, err
	}
	return s.convertResults(results), nil
}

// SearchMemories 语义搜索记忆（作用域感知）
// wxID 和 chatRoomID 共同决定搜索范围：
//   - wxID 有值, chatRoomID 为空 → 只搜索该用户的全局个人记忆
//   - wxID 有值, chatRoomID 有值 → 只搜索该用户在该群的群内个人记忆
//   - wxID 为空, chatRoomID 有值 → 只搜索该群的群级别记忆
func (s *VectorStoreService) SearchMemories(ctx context.Context, robotCode string, query, wxID, chatRoomID string, topK int) ([]ai.VectorSearchResult, error) {
	vector, err := s.embedding.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	var conditions []*pb.Condition
	if robotCode != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("robot_code", robotCode))
	}
	if wxID != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("contact_wxid", wxID))
	}
	// 始终过滤 chat_room_id：空字符串精确匹配全局记忆，非空匹配特定群
	conditions = append(conditions, qdrantx.BuildMatchFilter("chat_room_id", chatRoomID))

	filter := &pb.Filter{Must: conditions}

	results, err := s.qdrant.Search(ctx, qdrantx.CollectionMemories, vector, uint64(topK), filter)
	if err != nil {
		return nil, err
	}
	return s.convertResults(results), nil
}

// SearchKnowledge 语义搜索知识库
func (s *VectorStoreService) SearchKnowledge(ctx context.Context, robotCode string, query, category string, topK int) ([]ai.VectorSearchResult, error) {
	vector, err := s.embedding.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	var conditions []*pb.Condition
	if robotCode != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("robot_code", robotCode))
	}
	if category != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("category", category))
	}

	var filter *pb.Filter
	if len(conditions) > 0 {
		filter = &pb.Filter{Must: conditions}
	}

	results, err := s.qdrant.Search(ctx, qdrantx.CollectionKnowledge, vector, uint64(topK), filter)
	if err != nil {
		return nil, err
	}
	return s.convertResults(results), nil
}

// SearchKnowledgeByCategories 按多个分类语义搜索知识库
func (s *VectorStoreService) SearchKnowledgeByCategories(ctx context.Context, robotCode string, query string, categories []string, topK int) ([]ai.VectorSearchResult, error) {
	vector, err := s.embedding.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	var conditions []*pb.Condition
	if robotCode != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("robot_code", robotCode))
	}
	if len(categories) > 0 {
		categoryConditions := make([]*pb.Condition, 0, len(categories))
		for _, cat := range categories {
			categoryConditions = append(categoryConditions, qdrantx.BuildMatchFilter("category", cat))
		}
		conditions = append(conditions, &pb.Condition{
			ConditionOneOf: &pb.Condition_Filter{
				Filter: &pb.Filter{Should: categoryConditions},
			},
		})
	}
	var filter *pb.Filter
	if len(conditions) > 0 {
		filter = &pb.Filter{Must: conditions}
	}
	results, err := s.qdrant.Search(ctx, qdrantx.CollectionKnowledge, vector, uint64(topK), filter)
	if err != nil {
		return nil, err
	}
	return s.convertResults(results), nil
}

// DeleteVectors 删除向量
func (s *VectorStoreService) DeleteVectors(ctx context.Context, collection string, ids []string) error {
	return s.qdrant.DeleteByIDs(ctx, collection, ids)
}

// IndexImageKnowledge 将图片向量化并存入 Qdrant image_knowledge 集合
func (s *VectorStoreService) IndexImageKnowledge(ctx context.Context, robotCode string, docID int64, imageURL, title, description, category string) (string, error) {
	if s.imageEmbedding == nil {
		return "", fmt.Errorf("image embedding service not configured")
	}

	vector, err := s.imageEmbedding.EmbedImage(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("embed image: %w", err)
	}

	id := s.qdrant.GenerateID()
	payload := map[string]*pb.Value{
		"robot_code":  qdrantx.NewPayloadValue(robotCode),
		"doc_id":      qdrantx.NewPayloadIntValue(docID),
		"image_url":   qdrantx.NewPayloadValue(imageURL),
		"title":       qdrantx.NewPayloadValue(title),
		"description": qdrantx.NewPayloadValue(description),
		"category":    qdrantx.NewPayloadValue(category),
	}

	if err := s.qdrant.Upsert(ctx, qdrantx.CollectionImageKnowledge, id, vector, payload); err != nil {
		return "", fmt.Errorf("upsert image knowledge vector: %w", err)
	}
	return id, nil
}

// SearchImageKnowledgeByText 以文搜图：用文本查询搜索图片知识库
func (s *VectorStoreService) SearchImageKnowledgeByText(ctx context.Context, robotCode, query, category string, topK int) ([]ai.VectorSearchResult, error) {
	if s.imageEmbedding == nil {
		return nil, fmt.Errorf("image embedding service not configured")
	}

	vector, err := s.imageEmbedding.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed text query for image search: %w", err)
	}

	filter := s.buildImageKnowledgeFilter(robotCode, category)

	results, err := s.qdrant.Search(ctx, qdrantx.CollectionImageKnowledge, vector, uint64(topK), filter)
	if err != nil {
		return nil, err
	}
	return s.convertResults(results), nil
}

// SearchImageKnowledgeByImage 以图搜图：用图片搜索相似图片
func (s *VectorStoreService) SearchImageKnowledgeByImage(ctx context.Context, robotCode, imageURL, category string, topK int) ([]ai.VectorSearchResult, error) {
	if s.imageEmbedding == nil {
		return nil, fmt.Errorf("image embedding service not configured")
	}

	vector, err := s.imageEmbedding.EmbedImage(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("embed image for image search: %w", err)
	}

	filter := s.buildImageKnowledgeFilter(robotCode, category)

	results, err := s.qdrant.Search(ctx, qdrantx.CollectionImageKnowledge, vector, uint64(topK), filter)
	if err != nil {
		return nil, err
	}
	return s.convertResults(results), nil
}

func (s *VectorStoreService) buildImageKnowledgeFilter(robotCode, category string) *pb.Filter {
	var conditions []*pb.Condition
	if robotCode != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("robot_code", robotCode))
	}
	if category != "" {
		conditions = append(conditions, qdrantx.BuildMatchFilter("category", category))
	}
	if len(conditions) > 0 {
		return &pb.Filter{Must: conditions}
	}
	return nil
}

func (s *VectorStoreService) convertResults(points []*pb.ScoredPoint) []ai.VectorSearchResult {
	results := make([]ai.VectorSearchResult, 0, len(points))
	for _, p := range points {
		result := ai.VectorSearchResult{
			Score:   p.GetScore(),
			Payload: make(map[string]string),
		}
		if pid := p.GetId(); pid != nil {
			result.ID = pid.GetUuid()
		}
		for k, v := range p.GetPayload() {
			switch val := v.GetKind().(type) {
			case *pb.Value_StringValue:
				result.Payload[k] = val.StringValue
			case *pb.Value_IntegerValue:
				result.Payload[k] = strconv.FormatInt(val.IntegerValue, 10)
			case *pb.Value_DoubleValue:
				result.Payload[k] = strconv.FormatFloat(val.DoubleValue, 'f', -1, 64)
			}
		}
		results = append(results, result)
	}
	return results
}
