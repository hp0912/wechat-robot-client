package pkg

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type MimoTTSConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

type mimoDeltaAudio struct {
	Audio struct {
		Data string `json:"data"`
	} `json:"audio"`
}

// MimoTTSSubmit sends a streaming TTS request to the mimo API and returns WAV audio bytes.
func MimoTTSSubmit(config *MimoTTSConfig, text, voice string) ([]byte, error) {
	client := openai.NewClient(
		option.WithAPIKey(config.APIKey),
		option.WithBaseURL(config.BaseURL),
	)

	stream := client.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModel(config.Model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.AssistantMessage(text),
		},
		Audio: openai.ChatCompletionAudioParam{
			Format: openai.ChatCompletionAudioParamFormatPcm16,
			Voice: openai.ChatCompletionAudioParamVoiceUnion{
				OfChatCompletionAudioVoiceString2: openai.String(voice),
			},
		},
	})

	var pcmBuf bytes.Buffer
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}
		var da mimoDeltaAudio
		if err := json.Unmarshal([]byte(chunk.Choices[0].Delta.RawJSON()), &da); err != nil || da.Audio.Data == "" {
			continue
		}
		pcmBytes, err := base64.StdEncoding.DecodeString(da.Audio.Data)
		if err != nil {
			return nil, fmt.Errorf("解码 mimo 音频数据失败: %w", err)
		}
		pcmBuf.Write(pcmBytes)
	}
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("mimo TTS 流式请求失败: %w", err)
	}

	return pcm16LEToWAV(pcmBuf.Bytes(), 24000, 1), nil
}

// pcm16LEToWAV wraps raw PCM16LE samples with a standard WAV header.
func pcm16LEToWAV(pcm []byte, sampleRate uint32, channels uint16) []byte {
	var buf bytes.Buffer
	dataSize := uint32(len(pcm))
	byteRate := sampleRate * uint32(channels) * 2
	blockAlign := channels * 2

	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, 36+dataSize)
	buf.WriteString("WAVE")

	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, channels)
	binary.Write(&buf, binary.LittleEndian, sampleRate)
	binary.Write(&buf, binary.LittleEndian, byteRate)
	binary.Write(&buf, binary.LittleEndian, blockAlign)
	binary.Write(&buf, binary.LittleEndian, uint16(16))

	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, dataSize)
	buf.Write(pcm)

	return buf.Bytes()
}
