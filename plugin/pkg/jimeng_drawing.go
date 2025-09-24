package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type JimengRequest struct {
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	SampleStrength float64 `json:"sample_strength"`
	ResolutionType string  `json:"resolution_type"`
}

type JimengConfig struct {
	BaseURL   string   `json:"base_url"`
	SessionID []string `json:"sessionid"`
	JimengRequest
}

type JimengResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL string `json:"url"`
	} `json:"data"`
}

func JimengDrawing(config *JimengConfig) (string, error) {
	if config.Prompt == "" {
		return "", fmt.Errorf("绘图提示词为空")
	}
	if len(config.SessionID) == 0 {
		return "", fmt.Errorf("未找到绘图密钥")
	}
	// 设置默认值
	if config.Model == "" {
		config.Model = "jimeng-3.0"
	}
	if config.Width == 0 {
		config.Width = 1024
	}
	if config.Height == 0 {
		config.Height = 1024
	}
	if config.SampleStrength == 0 {
		config.SampleStrength = 0.5
	}
	sessionID := strings.Join(config.SessionID, ",")
	// 准备请求体
	requestBody, err := json.Marshal(config.JimengRequest)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %v", err)
	}
	// 创建HTTP请求
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/images/generations", config.BaseURL), bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionID)
	// 发送请求
	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API请求失败，状态码 %d: %s", resp.StatusCode, string(body))
	}
	// 解析响应
	var jimengResp JimengResponse
	if err := json.Unmarshal(body, &jimengResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	// 检查是否有生成的图片
	if len(jimengResp.Data) == 0 {
		return "", fmt.Errorf("未生成任何图片")
	}
	imageUrls := ""
	for index, data := range jimengResp.Data {
		if index > 0 {
			imageUrls += "\n"
		}
		imageUrls += data.URL
	}
	return imageUrls, nil
}
