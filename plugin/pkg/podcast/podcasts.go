package podcast

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type PodcastSecrets struct {
	AppID       string `json:"app_id"`
	AccessToken string `json:"access_token"`
	ResourceID  string `json:"resource_id"`
}

type PodcastConfig struct {
	Action       int         `json:"action"`
	PromptText   string      `json:"prompt_text,omitempty"`
	InputText    string      `json:"input_text,omitempty"`
	InputInfo    InputInfo   `json:"input_info,omitempty"`
	NLPTexts     []NLPText   `json:"nlp_texts,omitempty"`
	AudioConfig  AudioConfig `json:"audio_config,omitempty"`
	SpeakerInfo  SpeakerInfo `json:"speaker_info,omitempty"`
	RetryInfo    RetryInfo   `json:"retry_info,omitempty"`
	InputID      string      `json:"input_id,omitempty"`
	UseHeadMusic bool        `json:"use_head_music,omitempty"`
	UseTailMusic bool        `json:"use_tail_music,omitempty"`
	OnlyNLPText  bool        `json:"only_nlp_text,omitempty"`
}

type InputInfo struct {
	InputURL       string `json:"input_url,omitempty"`
	OnlyNLPText    bool   `json:"only_nlp_text,omitempty"`    // 只输出播客轮次文本列表，没有音频
	ReturnAudioURL bool   `json:"return_audio_url,omitempty"` // 返回可下载的完整播客音频链接，有效期 1h
}

type NLPText struct {
	Text    string `json:"text,omitempty"`
	Speaker string `json:"speaker,omitempty"`
}

type AudioConfig struct {
	Format     string `json:"format,omitempty"`      // mp3/ogg_opus/pcm/aac
	SampleRate int    `json:"sample_rate,omitempty"` // [16000, 24000, 48000]
	SpeechRate int    `json:"speech_rate,omitempty"` // 语速，取值范围[-50,100]，100代表2.0倍速，-50代表0.5倍数
}

type SpeakerInfo struct {
	RandomOrder bool     `json:"random_order,omitempty"` // 2发音人是否随机顺序开始，默认是
	Speakers    []string `json:"speakers,omitempty"`     // 播客发音人, 只能选择 2 发音人
}

type RetryInfo struct {
	RetryTaskID         string `json:"retry_task_id,omitempty"`          // 前一个没获取完整的播客记录的 task_id(第一次StartSession使用的 session_id就是任务的 task_id)
	LastFinishedRoundID int    `json:"last_finished_round_id,omitempty"` // 前一个获取完整的播客记录的轮次 id
}

type PodcastRoundStart struct {
	TextType string `json:"text_type"`
	Text     string `json:"text"`
	Speaker  string `json:"speaker"`
	RoundID  int    `json:"round_id"`
}

type PodcastEnd struct {
	MetaInfo MetaInfo `json:"meta_info"`
}

type MetaInfo struct {
	AudioURL     string       `json:"audio_url"`
	InputMetrics InputMetrics `json:"input_metrics"`
}

type InputMetrics struct {
	OriginInputTextLength int  `json:"origin_input_text_length"`
	InputTextLength       int  `json:"input_text_length"`
	InputTextTruncated    bool `json:"input_text_truncated"`
}

func Podcast(secrets PodcastSecrets, config PodcastConfig) (string, error) {
	if secrets.AppID == "" || secrets.AccessToken == "" {
		return "", fmt.Errorf("app_id and access_token are required")
	}
	if secrets.ResourceID == "" {
		secrets.ResourceID = "volc.service_type.10050"
	}

	config.InputID = uuid.New().String()
	config.UseHeadMusic = true
	config.UseTailMusic = true
	config.AudioConfig.Format = "mp3"
	config.AudioConfig.SampleRate = 24000
	config.InputInfo.ReturnAudioURL = true

	header := http.Header{}
	header.Set("X-Api-App-Id", secrets.AppID)
	header.Set("X-Api-App-Key", "aGjiRDfUWi")
	header.Set("X-Api-Access-Key", secrets.AccessToken)
	header.Set("X-Api-Resource-Id", secrets.ResourceID)
	header.Set("X-Api-Connect-Id", uuid.New().String())

	var (
		isPodcastRoundEnd = false              // 标志当前轮是否结束
		audioReceived     = false              // 标志是否收到音频数据
		lastRoundID       = -1                 // 上一轮的轮次ID
		taskID            = ""                 // 任务ID
		retryNum          = 5                  // 重试次数
		podcastAudio      = make([]byte, 0)    // 整个播客的音频数据
		audio             = make([]byte, 0)    // 当前轮的音频数据
		currentRound      = 0                  // 当前轮次ID
		podcastTexts      = make([]NLPText, 0) // 播客轮次文本列表
	)

	// 建立WebSocket连接	client <---> server
	conn, r, err := websocket.DefaultDialer.DialContext(context.Background(), "wss://openspeech.bytedance.com/api/v3/sami/podcasttts", header)
	if err != nil {
		return "", err
	}
	defer func() {
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Error during WebSocket close: %v\n", err)
		}
		err = conn.Close()
		if err != nil {
			log.Printf("Error during WebSocket close: %v\n", err)
		}
	}()

	log.Println("Connection established, Logid: ", r.Header.Get("x-tt-logid"))

	for retryNum > 0 {
		if !isPodcastRoundEnd {
			config.RetryInfo.RetryTaskID = taskID
			config.RetryInfo.LastFinishedRoundID = lastRoundID
		}

		// Start connection [event=1] --> server
		if err := StartConnection(conn); err != nil {
			return "", err
		}

		// Connection started [event=50] <-- server
		_, err = WaitForEvent(conn, MsgTypeFullServerResponse, EventType_ConnectionStarted)
		if err != nil {
			return "", err
		}

		sessionID := uuid.New().String()
		if taskID == "" {
			taskID = sessionID
		}

		payload, err := json.Marshal(&config)
		if err != nil {
			return "", err
		}

		// Start session [event=100] --> server
		if err := StartSession(conn, payload, sessionID); err != nil {
			return "", err
		}

		// Session started [event=150] <-- server
		_, err = WaitForEvent(conn, MsgTypeFullServerResponse, EventType_SessionStarted)
		if err != nil {
			return "", err
		}

		// Finish session [event=102] --> server
		if err := FinishSession(conn, sessionID); err != nil {
			return "", err
		}

		for {
			var msg *Message
			// 接收响应内容
			if msg, err = ReceiveMessage(conn); err != nil {
				return "", err
			}
			switch msg.MsgType {
			// 音频数据块
			case MsgTypeAudioOnlyServer:
				// 音频数据块
				if msg.EventType == EventType_PodcastRoundResponse {
					if !audioReceived && len(audio) > 0 {
						audioReceived = true
					}
					audio = append(audio, msg.Payload...)
				}
				// 错误信息
			case MsgTypeError:
				return "", fmt.Errorf("收到错误信息:%s", string(msg.Payload))
				// 其他消息类型
			case MsgTypeFullServerResponse:
				switch msg.EventType {
				// 播客开始
				case EventType_PodcastRoundStart:
					var data PodcastRoundStart
					if err := json.Unmarshal(msg.Payload, &data); err != nil {
						return "", fmt.Errorf("反序列化失败: %v", err)
					}
					if config.InputInfo.OnlyNLPText {
						podcastTexts = append(podcastTexts, NLPText{
							Speaker: data.Speaker,
							Text:    data.Text,
						})
					}
					currentRound = data.RoundID
					isPodcastRoundEnd = false
				// 播客结束
				case EventType_PodcastRoundEnd:
					var data struct {
						IsError  bool   `json:"is_error"`
						ErrorMsg string `json:"error_msg"`
					}
					if err := json.Unmarshal(msg.Payload, &data); err != nil {
						return "", fmt.Errorf("反序列化失败: %v", err)
					}
					if data.IsError {
						return "", fmt.Errorf("播客round结束, 有错误发生%s", data.ErrorMsg)
					}
					isPodcastRoundEnd = true
					lastRoundID = currentRound
					log.Printf("第 %d 轮结束\n", lastRoundID)
					if len(audio) > 0 {
						podcastAudio = append(podcastAudio, audio...)
						audio = make([]byte, 0)
					}
				case EventType_PodcastEnd:
					var data PodcastEnd
					if err := json.Unmarshal(msg.Payload, &data); err != nil {
						return "", fmt.Errorf("反序列化失败: %v", err)
					}
					if data.MetaInfo.AudioURL != "" {
						return data.MetaInfo.AudioURL, nil
					}
					log.Printf("播客结束: %v\n", data)
					continue
				}
			}
			// 会话结束
			if msg.EventType == EventType_SessionFinished {
				break
			}
		}

		if !audioReceived && !config.InputInfo.OnlyNLPText {
			return "", fmt.Errorf("未接收到音频数据内容")
		}

		// 保持连接，方便下次请求
		if err := FinishConnection(conn); err != nil {
			return "", fmt.Errorf("结束连接失败: %v", err)
		}

		_, err = WaitForEvent(conn, MsgTypeFullServerResponse, EventType_ConnectionFinished)
		if err != nil {
			return "", fmt.Errorf("等待连接结束事件失败: %v", err)
		}

		// 播客结束, 保存最终音频文件
		if isPodcastRoundEnd {
			if len(podcastAudio) > 0 {
				//
			}
			if len(podcastTexts) > 0 && config.InputInfo.OnlyNLPText {
				//
			}
			break
		} else {
			log.Printf("播客未结束，进入第 %d 轮\n", lastRoundID)
			retryNum--
			time.Sleep(1 * time.Second)
		}
	}

	return "", fmt.Errorf("将二进制流转播客敬请期待")
}
