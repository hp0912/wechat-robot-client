package service

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/redis/go-redis/v9"

	"wechat-robot-client/vars"
)

const (
	embeddingCacheDuration = 24 * time.Hour
	embeddingCachePrefix   = "emb:"
)

// EmbeddingService 向量化服务
type EmbeddingService struct {
	client    *openai.Client
	model     openai.EmbeddingModel
	dimension int
}

// NewEmbeddingService 创建向量化服务
func NewEmbeddingService(baseURL, apiKey, model string, dimension int) *EmbeddingService {
	embModel := openai.EmbeddingModel(model)
	if model == "" {
		embModel = openai.EmbeddingModelTextEmbedding3Small
	}
	if dimension <= 0 {
		dimension = 2048
	}
	client := newOpenAIClient(apiKey, baseURL)
	return &EmbeddingService{
		client:    &client,
		model:     embModel,
		dimension: dimension,
	}
}

// Embed 将单条文本转为向量
func (s *EmbeddingService) Embed(ctx context.Context, text string) ([]float32, error) {
	if cached, err := s.getFromCache(ctx, text); err == nil && cached != nil {
		return cached, nil
	}

	resp, err := s.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Model:      s.model,
		Dimensions: openai.Int(int64(s.dimension)),
	})
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	vector := float64SliceToFloat32(resp.Data[0].Embedding)
	s.setCache(ctx, text, vector)
	return vector, nil
}

// EmbedBatch 批量向量化
func (s *EmbeddingService) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	resp, err := s.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: texts,
		},
		Model:      s.model,
		Dimensions: openai.Int(int64(s.dimension)),
	})
	if err != nil {
		return nil, fmt.Errorf("batch embedding failed: %w", err)
	}

	results := make([][]float32, len(texts))
	for i, data := range resp.Data {
		index := i
		if data.Index >= 0 && int(data.Index) < len(results) {
			index = int(data.Index)
		}
		results[index] = float64SliceToFloat32(data.Embedding)
	}
	return results, nil
}

func float64SliceToFloat32(data []float64) []float32 {
	result := make([]float32, len(data))
	for i, v := range data {
		result[i] = float32(v)
	}
	return result
}

func (s *EmbeddingService) cacheKey(text string) string {
	hash := sha256.Sum256([]byte(string(s.model) + ":" + text))
	return embeddingCachePrefix + hex.EncodeToString(hash[:16])
}

func (s *EmbeddingService) getFromCache(ctx context.Context, text string) ([]float32, error) {
	if vars.RedisClient == nil {
		return nil, nil
	}
	key := s.cacheKey(text)
	data, err := vars.RedisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return bytesToFloat32Slice(data), nil
}

func (s *EmbeddingService) setCache(ctx context.Context, text string, vector []float32) {
	if vars.RedisClient == nil {
		return
	}
	key := s.cacheKey(text)
	vars.RedisClient.Set(ctx, key, float32SliceToBytes(vector), embeddingCacheDuration)
}

func float32SliceToBytes(data []float32) []byte {
	buf := make([]byte, len(data)*4)
	for i, v := range data {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

func bytesToFloat32Slice(data []byte) []float32 {
	if len(data)%4 != 0 {
		return nil
	}
	result := make([]float32, len(data)/4)
	for i := range result {
		result[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}
	return result
}
