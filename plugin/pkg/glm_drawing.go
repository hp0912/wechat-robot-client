package pkg

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type GLMConfig struct {
	ApiKey  string  `json:"api_key"`
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Quality *string `json:"quality"`
	Size    *string `json:"size"`
	UserID  *string `json:"user_id"`
}

type DataItem struct {
	Url string `json:"url"`
}

type ContentFilter struct {
	Role  string `json:"role"`
	Level int    `json:"level"`
}

type GLMResponse struct {
	Created       int64           `json:"created"`        // 请求创建时间，是以秒为单位的Unix时间戳
	Data          []DataItem      `json:"data"`           // 数组，包含生成的图片 URL。目前数组中只包含一张图片
	ContentFilter []ContentFilter `json:"content_filter"` // 返回内容安全的相关信息
}

// GLMDrawing 智谱绘图
func GLMDrawing(config *GLMConfig) (string, error) {
	var respData GLMResponse
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.ApiKey)).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]any{
			"model":   config.Model,
			"prompt":  config.Prompt,
			"quality": config.Quality,
			"size":    config.Size,
			"user_id": config.UserID,
		}).
		SetResult(&respData).
		Post("https://open.bigmodel.cn/api/paas/v4/images/generations")
	if err != nil {
		return "", fmt.Errorf("调用绘图接口失败: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("调用绘图接口失败，返回状态码不是 200: %d", resp.StatusCode())
	}
	if len(respData.Data) == 0 || respData.Data[0].Url == "" {
		return "", fmt.Errorf("调用绘图接口失败: 图片地址为空")
	}
	return respData.Data[0].Url, nil
}
