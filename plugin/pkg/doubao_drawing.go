package pkg

import (
	"context"
	"fmt"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type DoubaoConfig struct {
	ApiKey         string  `json:"api_key"`
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt"`
	Image          string  `json:"image"`
	ResponseFormat string  `json:"response_format"`
	Size           string  `json:"size"`
	Seed           int64   `json:"seed"`
	GuidanceScale  float64 `json:"guidance_scale"`
	Watermark      bool    `json:"watermark"`
}

// DoubaoDrawing 豆包绘图
func DoubaoDrawing(config *DoubaoConfig) (string, error) {
	client := arkruntime.NewClientWithApiKey(config.ApiKey)
	ctx := context.Background()
	format := string(model.GenerateImagesResponseFormatURL)

	generateReq := model.GenerateImagesRequest{
		Model:          config.Model,
		Prompt:         config.Prompt,
		ResponseFormat: &format,
		Watermark:      &config.Watermark,
	}
	if config.Image != "" {
		generateReq.Image = &config.Image
	}
	if config.ResponseFormat != "" {
		generateReq.ResponseFormat = &config.ResponseFormat
	}
	if config.Size != "" {
		generateReq.Size = &config.Size
	}
	if config.Seed != 0 {
		generateReq.Seed = &config.Seed
	}
	if config.GuidanceScale != 0 {
		generateReq.GuidanceScale = &config.GuidanceScale
	}

	imagesResponse, err := client.GenerateImages(ctx, generateReq)
	if err != nil {
		return "", fmt.Errorf("generate images error: %v", err)
	}
	if imagesResponse.Error != nil {
		return "", fmt.Errorf("generate images error: %s", imagesResponse.Error.Message)
	}
	if len(imagesResponse.Data) == 0 {
		return "", fmt.Errorf("no images generated")
	}
	if imagesResponse.Data[0].Url == nil {
		return "", fmt.Errorf("no image URL found")
	}

	return *imagesResponse.Data[0].Url, nil
}
