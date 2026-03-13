package service

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"wechat-robot-client/utils"
	"wechat-robot-client/vars"

	"github.com/redis/go-redis/v9"
	"github.com/sashabaranov/go-openai"
)

const (
	embeddingDimensions    = 1536
	embeddingCacheDuration = 24 * time.Hour
	embeddingCachePrefix   = "emb:"
)

// EmbeddingService 向量化服务
type EmbeddingService struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

// NewEmbeddingService 创建向量化服务
func NewEmbeddingService(baseURL, apiKey, model string) *EmbeddingService {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = utils.NormalizeAIBaseURL(baseURL)
	embModel := openai.EmbeddingModel(model)
	if model == "" {
		embModel = openai.SmallEmbedding3
	}
	return &EmbeddingService{
		client: openai.NewClientWithConfig(config),
		model:  embModel,
	}
}

// Embed 将单条文本转为向量
func (s *EmbeddingService) Embed(ctx context.Context, text string) ([]float32, error) {
	if cached, err := s.getFromCache(ctx, text); err == nil && cached != nil {
		return cached, nil
	}

	resp, err := s.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input:      []string{text},
		Model:      s.model,
		Dimensions: embeddingDimensions,
	})
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	vector := resp.Data[0].Embedding
	s.setCache(ctx, text, vector)
	return vector, nil
}

// EmbedBatch 批量向量化
func (s *EmbeddingService) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	resp, err := s.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input:      texts,
		Model:      s.model,
		Dimensions: embeddingDimensions,
	})
	if err != nil {
		return nil, fmt.Errorf("batch embedding failed: %w", err)
	}

	results := make([][]float32, len(texts))
	for i, data := range resp.Data {
		results[i] = data.Embedding
	}
	return results, nil
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
