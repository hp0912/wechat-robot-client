package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DoubaoLongTextTTSRequest struct {
	AppID            string   `json:"appid"`
	ReqID            string   `json:"reqid"`  // Request ID，不可重复，长度20～64，建议使用uuid
	Text             string   `json:"text"`   // 合成文本，长度小于10万字符，支持SSML
	Format           string   `json:"format"` //输出音频格式，支持pcm/wav/mp3/ogg_opus
	VoiceType        string   `json:"voice_type"`
	Voice            *string  `json:"voice"`
	Language         *string  `json:"language"`
	SampleRate       *int     `json:"sample_rate"`       // 采样率，默认为24000
	Volume           *float64 `json:"volume"`            // 音量，范围0.1～3，默认为1
	Speed            *float64 `json:"speed"`             // 语速，范围0.2～3，默认为1
	Pitch            *float64 `json:"pitch"`             // 语调，范围0.1～3，默认为1
	EnableSubtitle   *int     `json:"enable_subtitle"`   // 是否开启字幕时间戳，0表示不开启，1表示开启句级别字幕时间戳，2表示开启字词级别时间戳，3表示开启音素级别时间戳
	SentenceInterval *int     `json:"sentence_interval"` // 句间停顿，单位毫秒，范围0～3000，默认为预测值
	Style            *string  `json:"style"`             // 指定情感，“情感预测版”默认为预测值，“普通版”默认为音色默认值
	CallbackURL      *string  `json:"callback_url"`
}

type DoubaoLongTextTTSConfig struct {
	BaseURL     string `json:"base_url"`
	AccessToken string `json:"access_token"`
	TaskID      string `json:"task_id"` // 用于查询任务状态
	DoubaoLongTextTTSRequest
}

type DoubaoLongTextTTSResponse struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	ReqID      string `json:"reqid"`
	TaskID     string `json:"task_id"`
	TaskStatus int    `json:"task_status"`
	TextLength int    `json:"text_length"`
}

type Phoneme struct {
	Ph    string `form:"ph" json:"ph"`
	Begin int    `form:"begin" json:"begin"`
	End   int    `form:"end" json:"end"`
}

type Word struct {
	Text     string    `form:"text" json:"text"`
	Begin    int       `form:"begin" json:"begin"`
	End      int       `form:"end" json:"end"`
	Phonemes []Phoneme `form:"phonemes" json:"phonemes"`
}

type Sentence struct {
	Text        string `form:"text" json:"text"`
	OriginText  string `form:"origin_text" json:"origin_text"`
	ParagraphNo int    `form:"paragraph_no" json:"paragraph_no"`
	BeginTime   int    `form:"begin_time" json:"begin_time"`
	EndTime     int    `form:"end_time" json:"end_time"`
	Emotion     string `form:"emotion" json:"emotion"`
	Words       []Word `form:"words" json:"words"`
}

type DoubaoLongTextTTSQueryResponse struct {
	Code          int        `form:"code" json:"code"`
	Message       string     `form:"message" json:"message"`
	TaskID        string     `form:"task_id" json:"task_id"`
	TaskStatus    int        `form:"task_status" json:"task_status"`
	TextLength    int        `form:"text_length" json:"text_length"`
	AudioURL      string     `form:"audio_url" json:"audio_url"`
	URLExpireTime int        `form:"url_expire_time" json:"url_expire_time"`
	Sentences     []Sentence `form:"sentences" json:"sentences"`
}

func DoubaoLongTextTTSSubmit(config *DoubaoLongTextTTSConfig) (string, error) {
	if config.Text == "" {
		return "", fmt.Errorf("合成文本为空")
	}
	if config.AppID == "" {
		return "", fmt.Errorf("应用ID不能为空")
	}
	if config.AccessToken == "" {
		return "", fmt.Errorf("未找到语音合成密钥")
	}

	config.ReqID = uuid.NewString()
	if config.Format == "" {
		config.Format = "mp3" // 默认格式为mp3
	}
	if config.VoiceType == "" {
		config.VoiceType = "BV104_streaming" // 温柔淑女
	}
	// 准备请求体
	requestBody, err := json.Marshal(config.DoubaoLongTextTTSRequest)
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
	var ttsResp DoubaoLongTextTTSResponse
	if err := json.Unmarshal(body, &ttsResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	if ttsResp.Code != 0 || ttsResp.TaskStatus != 0 {
		return "", fmt.Errorf("合成失败: %s", ttsResp.Message)
	}
	return ttsResp.TaskID, nil
}

func DoubaoLongTextTTSQuery(config *DoubaoLongTextTTSConfig) (string, error) {
	if config.AppID == "" {
		return "", fmt.Errorf("应用ID不能为空")
	}
	if config.TaskID == "" {
		return "", fmt.Errorf("任务ID不能为空")
	}
	path, err := url.Parse(strings.Replace(config.BaseURL, "/submit", "/query", 1))
	if err != nil {
		return "", fmt.Errorf("解析BaseURL失败: %v", err)
	}
	params := url.Values{}
	params.Add("appid", config.AppID)
	params.Add("task_id", config.TaskID)
	path.RawQuery = params.Encode()
	req, err := http.NewRequest(http.MethodGet, path.String(), nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer; %s", config.AccessToken))
	// 发送请求
	client := &http.Client{Timeout: 60 * time.Second}
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
	var ttsResp DoubaoLongTextTTSQueryResponse
	if err := json.Unmarshal(body, &ttsResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	if ttsResp.Code != 0 || ttsResp.TaskStatus != 0 {
		return "", fmt.Errorf("合成失败: %s", ttsResp.Message)
	}
	return ttsResp.TaskID, nil
}
