package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type DoubaoTTSConfig struct {
	BaseURL     string `json:"base_url"`
	AccessToken string `json:"access_token"`
	DoubaoTTSRequest
}

type DoubaoTTSRequest struct {
	App     AppConfig     `json:"app"`
	User    UserConfig    `json:"user"`
	Audio   AudioConfig   `json:"audio"`
	Request RequestConfig `json:"request"`
}

type AppConfig struct {
	AppID   string `json:"appid"`
	Token   string `json:"token"`
	Cluster string `json:"cluster"`
}

type UserConfig struct {
	UID string `json:"uid"`
}

type AudioConfig struct {
	VoiceType       string  `json:"voice_type"`
	Encoding        string  `json:"encoding"`
	CompressionRate int     `json:"compression_rate"`
	Rate            int     `json:"rate"`
	SpeedRatio      float64 `json:"speed_ratio"`
	VolumeRatio     float64 `json:"volume_ratio"`
	PitchRatio      float64 `json:"pitch_ratio"`
	Emotion         string  `json:"emotion"`
	Language        string  `json:"language"`
}

type RequestConfig struct {
	ReqID           string `json:"reqid"`
	Text            string `json:"text"`
	TextType        string `json:"text_type"`
	Operation       string `json:"operation"`
	SilenceDuration string `json:"silence_duration"`
	WithFrontend    string `json:"with_frontend"`
	FrontendType    string `json:"frontend_type"`
	PureEnglishOpt  string `json:"pure_english_opt"`
}

type DoubaoTTSResponse struct {
	ReqID     string   `json:"reqid"`
	Code      int      `json:"code"`
	Operation string   `json:"operation"`
	Message   string   `json:"message"`
	Sequence  int      `json:"sequence"`
	Data      string   `json:"data"`
	Addition  Addition `json:"addition"`
}

type Addition struct {
	Description string `json:"description"`
	Duration    string `json:"duration"`
	Frontend    string `json:"frontend"`
}

func DoubaoTTSSubmit(config *DoubaoTTSConfig) (string, error) {
	if config.App.AppID == "" {
		return "", fmt.Errorf("应用ID不能为空")
	}
	if config.AccessToken == "" {
		return "", fmt.Errorf("未找到语音合成密钥")
	}

	config.App.Token = uuid.NewString()
	config.App.Cluster = "volcano_tts"
	config.User.UID = uuid.NewString()
	config.Audio.Encoding = "mp3"
	config.Request.ReqID = uuid.NewString()
	config.Request.Operation = "query"
	config.Request.TextType = "plain"

	// 准备请求体
	requestBody, err := json.Marshal(config.DoubaoTTSRequest)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %v", err)
	}
	// 创建HTTP请求
	req, err := http.NewRequest("POST", config.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer; %s", config.AccessToken))

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
	var ttsResp DoubaoTTSResponse
	if err := json.Unmarshal(body, &ttsResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	if ttsResp.Message != "Success" {
		return "", fmt.Errorf("合成失败: %s", ttsResp.Message)
	}
	return ttsResp.Data, nil
}
