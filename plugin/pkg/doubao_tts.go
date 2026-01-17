package pkg

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type DoubaoTTSConfig struct {
	URL           string        `json:"url"`
	RequestHeader RequestHeader `json:"request_header"`
	RequestBody   RequestBody   `json:"request_body"`
}

type RequestHeader struct {
	XApiAppID                        string `json:"X-Api-App-Id"`
	XApiAccessKey                    string `json:"X-Api-Access-Key"`
	XApiResourceID                   string `json:"X-Api-Resource-Id"`
	XApiRequestID                    string `json:"X-Api-Request-Id,omitempty"`
	XControlRequireUsageTokensReturn string `json:"X-Control-Require-Usage-Tokens-Return,omitempty"`
}

type RequestBody struct {
	User      User      `json:"user"`
	Namespace string    `json:"namespace,omitempty"`
	ReqParams ReqParams `json:"req_params"`
}

type User struct {
	UID string `json:"uid,omitempty"`
}

type ReqParams struct {
	Text        string      `json:"text"`
	Model       string      `json:"model"`
	Speaker     string      `json:"speaker"`
	AudioParams AudioParams `json:"audio_params"`
	XAdditions  Additions   `json:"x-additions"`
	Additions   string      `json:"additions,omitempty"`
}

type AudioParams struct {
	Format       string `json:"format,omitempty"`
	SampleRate   int    `json:"sample_rate,omitempty"`
	BitRate      int    `json:"bit_rate,omitempty"`
	Emotion      string `json:"emotion,omitempty"`
	EmotionScale int    `json:"emotion_scale,omitempty"`
	SpeechRate   int    `json:"speech_rate,omitempty"`
	LoudnessRate int    `json:"loudness_rate,omitempty"`
}

type Additions struct {
	SilenceDuration              int      `json:"silence_duration,omitempty"`
	EnableLanguageDetector       bool     `json:"enable_language_detector,omitempty"`
	DisableMarkdownFilter        bool     `json:"disable_markdown_filter,omitempty"`
	DisableEmojiFilter           bool     `json:"disable_emoji_filter,omitempty"`
	MuteCutRemainMs              string   `json:"mute_cut_remain_ms,omitempty"`
	EnableLatexTn                bool     `json:"enable_latex_tn,omitempty"`
	LatexParser                  string   `json:"latex_parser,omitempty"`
	MaxLengthToFilterParenthesis int      `json:"max_length_to_filter_parenthesis,omitempty"`
	ExplicitLanguage             string   `json:"explicit_language,omitempty"`
	ContextLanguage              string   `json:"context_language,omitempty"`
	UnsupportedCharRatioThresh   float64  `json:"unsupported_char_ratio_thresh,omitempty"`
	AigcWatermark                bool     `json:"aigc_watermark,omitempty"`
	ContextTexts                 []string `json:"context_texts,omitempty"`
}

type DoubaoTTSResponse struct {
	Code     int            `json:"code"`
	Message  string         `json:"message"`
	Data     string         `json:"data"`
	Sentence map[string]any `json:"sentence,omitempty"`
	Usage    map[string]any `json:"usage,omitempty"`
}

func DoubaoTTSSubmit(config *DoubaoTTSConfig) (string, error) {
	if config.URL == "" {
		return "", fmt.Errorf("语音合成地址不能为空")
	}
	if config.RequestHeader.XApiAppID == "" || config.RequestHeader.XApiAccessKey == "" || config.RequestHeader.XApiResourceID == "" {
		return "", fmt.Errorf("请求头参数不能为空")
	}

	config.RequestBody.User.UID = uuid.NewString()
	if config.RequestBody.ReqParams.Speaker == "" {
		config.RequestBody.ReqParams.Speaker = "zh_female_vv_uranus_bigtts"
	}
	config.RequestBody.ReqParams.AudioParams.Format = "mp3"
	config.RequestBody.ReqParams.AudioParams.SampleRate = 24000

	// 准备请求体
	requestBody, err := json.Marshal(config.RequestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %v", err)
	}
	// 创建HTTP请求
	req, err := http.NewRequest("POST", config.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-App-Id", config.RequestHeader.XApiAppID)
	req.Header.Set("X-Api-Access-Key", config.RequestHeader.XApiAccessKey)
	req.Header.Set("X-Api-Resource-Id", config.RequestHeader.XApiResourceID)
	if config.RequestHeader.XApiRequestID != "" {
		req.Header.Set("X-Api-Request-Id", config.RequestHeader.XApiRequestID)
	}
	if config.RequestHeader.XControlRequireUsageTokensReturn != "" {
		req.Header.Set("X-Control-Require-Usage-Tokens-Return", config.RequestHeader.XControlRequireUsageTokensReturn)
	}

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败，状态码 %d: %s", resp.StatusCode, string(body))
	}

	audioData := make([]byte, 0)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var ttsResp DoubaoTTSResponse
		if err := json.Unmarshal([]byte(line), &ttsResp); err != nil {
			return "", fmt.Errorf("解析响应失败: %v, 行内容: %s", err, line)
		}

		if ttsResp.Code == 0 && ttsResp.Data != "" {
			chunkAudio, err := base64.StdEncoding.DecodeString(ttsResp.Data)
			if err != nil {
				return "", fmt.Errorf("解码音频数据失败: %v", err)
			}
			audioData = append(audioData, chunkAudio...)
			continue
		}

		if ttsResp.Code == 0 && ttsResp.Sentence != nil {
			continue
		}

		// 处理结束标识
		if ttsResp.Code == 20000000 {
			// 合成成功结束
			break
		}

		if ttsResp.Code > 0 {
			return "", fmt.Errorf("合成失败，错误码: %d, 错误信息: %s", ttsResp.Code, ttsResp.Message)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取响应流失败: %v", err)
	}

	if len(audioData) == 0 {
		return "", fmt.Errorf("未接收到音频数据")
	}

	return base64.StdEncoding.EncodeToString(audioData), nil
}
