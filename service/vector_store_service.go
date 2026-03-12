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
	qdrant    *qdrantx.QdrantClient
	embedding *EmbeddingService
}

// NewVectorStoreService 创建向量存储服务
func NewVectorStoreService(qdrant *qdrantx.QdrantClient, embedding *EmbeddingService) *VectorStoreService {
	return &VectorStoreService{
		qdrant:    qdrant,
		embedding: embedding,
	}
}

// IndexMessage 将消息内容向量化并存入 Qdrant
func (s *VectorStoreService) IndexMessage(ctx context.Context, robotCode string, msgID int64, content, contactWxID, chatRoomID, senderWxID string, createdAt int64) (string, error) {
	vector, err := s.embedding.Embed(ctx, content)
	if err != nil {
		return "", fmt.Errorf("embed message: %w", err)
	}

	id := s.qdrant.GenerateID()
	payload := map[string]*pb.Value{
		"robot_code":   qdrantx.NewPayloadValue(robotCode),
		"msg_id":       qdrantx.NewPayloadIntValue(msgID),
		"content":      qdrantx.NewPayloadValue(content),
		"contact_wxid": qdrantx.NewPayloadValue(contactWxID),
		"chat_room_id": qdrantx.NewPayloadValue(chatRoomID),
		"sender_wxid":  qdrantx.NewPayloadValue(senderWxID),
		"created_at":   qdrantx.NewPayloadIntValue(createdAt),
	}

	if err := s.qdrant.Upsert(ctx, qdrantx.CollectionMessages, id, vector, payload); err != nil {
		return "", fmt.Errorf("upsert message vector: %w", err)
	}
	return id, nil
}

// IndexMemory 将记忆内容向量化并存入 Qdrant
func (s *VectorStoreService) IndexMemory(ctx context.Context, robotCode string, memoryID int64, content, contactWxID, memoryType, key string) (string, error) {
	vector, err := s.embedding.Embed(ctx, content)
	if err != nil {
		return "", fmt.Errorf("embed memory: %w", err)
	}

	id := s.qdrant.GenerateID()
	payload := map[string]*pb.Value{
		"robot_code":   qdrantx.NewPayloadValue(robotCode),
		"memory_id":    qdrantx.NewPayloadIntValue(memoryID),
		"content":      qdrantx.NewPayloadValue(content),
		"contact_wxid": qdrantx.NewPayloadValue(contactWxID),
		"type":         qdrantx.NewPayloadValue(memoryType),
		"key":          qdrantx.NewPayloadValue(key),
	}

	if err := s.qdrant.Upsert(ctx, qdrantx.CollectionMemories, id, vector, payload); err != nil {
		return "", fmt.Errorf("upsert memory vector: %w", err)
	}
	return id, nil
}

// IndexKnowledge 将知识库内容向量化并存入 Qdrant
func (s *VectorStoreService) IndexKnowledge(ctx context.Context, robotCode string, docID int64, content, title, category string) (string, error) {
	vector, err := s.embedding.Embed(ctx, content)
	if err != nil {
		return "", fmt.Errorf("embed knowledge: %w", err)
	}

	id := s.qdrant.GenerateID()
	payload := map[string]*pb.Value{
		"robot_code": qdrantx.NewPayloadValue(robotCode),
		"doc_id":     qdrantx.NewPayloadIntValue(docID),
		"content":    qdrantx.NewPayloadValue(content),
		"title":      qdrantx.NewPayloadValue(title),
		"category":   qdrantx.NewPayloadValue(category),
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

// SearchMemories 语义搜索记忆
func (s *VectorStoreService) SearchMemories(ctx context.Context, robotCode string, query, contactWxID string, topK int) ([]ai.VectorSearchResult, error) {
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

	var filter *pb.Filter
	if len(conditions) > 0 {
		filter = &pb.Filter{Must: conditions}
	}

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

// DeleteVectors 删除向量
func (s *VectorStoreService) DeleteVectors(ctx context.Context, collection string, ids []string) error {
	return s.qdrant.DeleteByIDs(ctx, collection, ids)
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
