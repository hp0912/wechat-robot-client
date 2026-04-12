package qdrantx

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	pb "github.com/qdrant/go-client/qdrant"
)

const DefaultEmbeddingDimension uint64 = 1536
const (
	CollectionMessages       = "messages"
	CollectionMemories       = "memories"
	CollectionKnowledge      = "knowledge"
	CollectionImageKnowledge = "image_knowledge"
)

// QdrantClient 封装 Qdrant 客户端
type QdrantClient struct {
	client *pb.Client
}

// NewQdrantClient 创建 Qdrant 客户端
func NewQdrantClient(host string, port int, apiKey string) (*QdrantClient, error) {
	client, err := pb.NewClient(&pb.Config{
		Host:     host,
		Port:     port,
		APIKey:   apiKey,
		PoolSize: 3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}
	return &QdrantClient{client: client}, nil
}

// InitCollections 初始化文本相关集合（messages, memories, knowledge）
func (q *QdrantClient) InitCollections(ctx context.Context, dimension uint64) error {
	collections := []string{CollectionMessages, CollectionMemories, CollectionKnowledge}
	for _, collectionName := range collections {
		if err := q.InitCollection(ctx, collectionName, dimension); err != nil {
			return err
		}
	}
	return nil
}

// InitCollection 初始化单个集合
func (q *QdrantClient) InitCollection(ctx context.Context, name string, dimension uint64) error {
	exists, err := q.client.CollectionExists(ctx, name)
	if err != nil {
		return fmt.Errorf("check collection %s: %w", name, err)
	}
	if exists {
		return nil
	}
	err = q.client.CreateCollection(ctx, &pb.CreateCollection{
		CollectionName: name,
		VectorsConfig: pb.NewVectorsConfig(&pb.VectorParams{
			Size:     dimension,
			Distance: pb.Distance_Cosine,
		}),
	})
	if err != nil {
		return fmt.Errorf("create collection %s: %w", name, err)
	}
	if err := q.createPayloadIndexes(ctx, name); err != nil {
		log.Printf("[Qdrant] 创建索引失败 %s: %v", name, err)
	}
	log.Printf("[Qdrant] 创建集合 %s 成功", name)
	return nil
}

func (q *QdrantClient) createPayloadIndexes(ctx context.Context, collection string) error {
	indexes := map[string][]string{
		CollectionMessages:       {"robot_code", "contact_wxid", "chat_room_id", "chat_room_member_wxid"},
		CollectionMemories:       {"robot_code", "contact_wxid", "chat_room_id", "category"},
		CollectionKnowledge:      {"robot_code", "category", "title"},
		CollectionImageKnowledge: {"robot_code", "category", "title"},
	}
	fields, ok := indexes[collection]
	if !ok {
		return nil
	}
	for _, field := range fields {
		_, err := q.client.GetPointsClient().CreateFieldIndex(ctx, &pb.CreateFieldIndexCollection{
			CollectionName: collection,
			FieldName:      field,
			FieldType:      pb.FieldType_FieldTypeKeyword.Enum(),
		})
		if err != nil {
			return fmt.Errorf("create index %s.%s: %w", collection, field, err)
		}
	}
	return nil
}

// Upsert 插入或更新向量点
func (q *QdrantClient) Upsert(ctx context.Context, collection string, id string, vector []float32, payload map[string]*pb.Value) error {
	pointID := q.toPointID(id)
	_, err := q.client.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: collection,
		Points: []*pb.PointStruct{
			{
				Id:      pointID,
				Vectors: pb.NewVectors(vector...),
				Payload: payload,
			},
		},
	})
	return err
}

// UpsertBatch 批量插入向量点
func (q *QdrantClient) UpsertBatch(ctx context.Context, collection string, points []*pb.PointStruct) error {
	if len(points) == 0 {
		return nil
	}
	_, err := q.client.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: collection,
		Points:         points,
	})
	return err
}

// Search 语义搜索
func (q *QdrantClient) Search(ctx context.Context, collection string, vector []float32, topK uint64, filter *pb.Filter) ([]*pb.ScoredPoint, error) {
	results, err := q.client.Query(ctx, &pb.QueryPoints{
		CollectionName: collection,
		Query:          pb.NewQueryDense(vector),
		Limit:          &topK,
		Filter:         filter,
		WithPayload:    &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		return nil, fmt.Errorf("search %s: %w", collection, err)
	}
	return results, nil
}

// DeleteByIDs 根据 ID 列表删除向量点
func (q *QdrantClient) DeleteByIDs(ctx context.Context, collection string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	pointIDs := make([]*pb.PointId, len(ids))
	for i, id := range ids {
		pointIDs[i] = q.toPointID(id)
	}
	_, err := q.client.Delete(ctx, &pb.DeletePoints{
		CollectionName: collection,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Points{
				Points: &pb.PointsIdsList{Ids: pointIDs},
			},
		},
	})
	return err
}

// DeleteByFilter 根据过滤条件删除向量点
func (q *QdrantClient) DeleteByFilter(ctx context.Context, collection string, filter *pb.Filter) error {
	_, err := q.client.Delete(ctx, &pb.DeletePoints{
		CollectionName: collection,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: filter,
			},
		},
	})
	return err
}

// GenerateID 生成唯一 ID
func (q *QdrantClient) GenerateID() string {
	return uuid.New().String()
}

func (q *QdrantClient) toPointID(id string) *pb.PointId {
	return &pb.PointId{
		PointIdOptions: &pb.PointId_Uuid{Uuid: id},
	}
}

// Close 关闭连接
func (q *QdrantClient) Close() error {
	return q.client.Close()
}

// BuildMatchFilter 构建过滤条件辅助方法
func BuildMatchFilter(field, value string) *pb.Condition {
	return &pb.Condition{
		ConditionOneOf: &pb.Condition_Field{
			Field: &pb.FieldCondition{
				Key: field,
				Match: &pb.Match{
					MatchValue: &pb.Match_Keyword{Keyword: value},
				},
			},
		},
	}
}

// NewPayloadValue 创建 payload 字符串值
func NewPayloadValue(value string) *pb.Value {
	return &pb.Value{
		Kind: &pb.Value_StringValue{StringValue: value},
	}
}

// NewPayloadIntValue 创建 payload 整数值
func NewPayloadIntValue(value int64) *pb.Value {
	return &pb.Value{
		Kind: &pb.Value_IntegerValue{IntegerValue: value},
	}
}
