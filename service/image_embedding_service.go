package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ImageEmbeddingService 图片向量化服务（支持多模态 embedding API）
type ImageEmbeddingService struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// NewImageEmbeddingService 创建图片向量化服务
func NewImageEmbeddingService(baseURL, apiKey, model string) *ImageEmbeddingService {
	return &ImageEmbeddingService{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

type multiModalInput struct {
	Text  string `json:"text,omitempty"`
	Image string `json:"image,omitempty"`
}

type imageEmbeddingRequest struct {
	Model string            `json:"model"`
	Input []multiModalInput `json:"input"`
}

type imageEmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type imageEmbeddingResponse struct {
	Data []imageEmbeddingData `json:"data"`
}

// EmbedImage 将图片 URL 或 base64 转为向量
func (s *ImageEmbeddingService) EmbedImage(ctx context.Context, imageURL string) ([]float32, error) {
	return s.embed(ctx, []multiModalInput{{Image: imageURL}})
}

// EmbedText 将文本转为向量（用于以文搜图，文本与图片在同一向量空间）
func (s *ImageEmbeddingService) EmbedText(ctx context.Context, text string) ([]float32, error) {
	return s.embed(ctx, []multiModalInput{{Text: text}})
}

func (s *ImageEmbeddingService) embed(ctx context.Context, input []multiModalInput) ([]float32, error) {
	reqBody := imageEmbeddingRequest{
		Model: s.model,
		Input: input,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := s.baseURL + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("image embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image embedding API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result imageEmbeddingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("empty image embedding response")
	}

	return result.Data[0].Embedding, nil
}
