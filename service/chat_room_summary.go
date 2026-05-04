package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	htmltemplate "html/template"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/openai/openai-go/v3"

	chatRoomSummaryTemplate "wechat-robot-client/pkg/templates/chatroomsummary"
)

const chatRoomSummaryMaxCompletionTokens = 4000

type chatRoomSummaryReport struct {
	Overall   string                    `json:"overall"`
	Topics    []chatRoomSummaryTopic    `json:"topics"`
	Resources []chatRoomSummaryResource `json:"resources"`
}

type chatRoomSummaryTopic struct {
	Title        string   `json:"title"`
	Heat         int      `json:"heat"`
	Participants []string `json:"participants"`
	StartTime    string   `json:"start_time"`
	EndTime      string   `json:"end_time"`
	Process      string   `json:"process"`
	Comment      string   `json:"comment"`
}

type chatRoomSummaryResource struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type chatRoomSummaryTemplateData struct {
	Title         string
	SummaryModel  string
	ChatRoomName  string
	GeneratedAt   string
	Overall       string
	TopicCount    int
	ResourceCount int
	MaxHeatText   string
	Topics        []chatRoomSummaryTopicView
	Resources     []chatRoomSummaryResourceView
}

type chatRoomSummaryTopicView struct {
	Number       string
	Title        string
	HeatIcon     string
	HeatLabel    string
	TimeRange    string
	Participants []string
	Process      string
	Comment      string
	Accent       string
	SpanClass    string
}

type chatRoomSummaryResourceView struct {
	Title       string
	URL         string
	Description string
	Icon        string
}

func (s *ChatRoomService) generateChatRoomSummaryReport(ctx context.Context, apiKey, baseURL, summaryModel, chatRoomName string, content []string) (*chatRoomSummaryReport, error) {
	client := newOpenAIClient(apiKey, baseURL)
	messages := buildChatRoomSummaryAIMessages(chatRoomName, strings.Join(content, "\n"))

	report, err := requestChatRoomSummaryReport(ctx, &client, summaryModel, messages, true)
	if err == nil {
		return report, nil
	}

	log.Printf("群聊记录结构化总结失败，尝试降级为普通 JSON 输出: %v", err)
	fallbackReport, fallbackErr := requestChatRoomSummaryReport(ctx, &client, summaryModel, messages, false)
	if fallbackErr != nil {
		return nil, fmt.Errorf("结构化总结失败: %w; 降级总结失败: %v", err, fallbackErr)
	}
	return fallbackReport, nil
}

func buildChatRoomSummaryAIMessages(chatRoomName, transcript string) []openai.ChatCompletionMessageParamUnion {
	prompt := `你是一个中文微信群聊总结助手。请从群聊记录中提取今日群聊报告，并且只返回一个严格 JSON 对象，不要返回 Markdown，不要使用代码块。

每一行代表一个人的发言，格式为：[time] {"nickname": "content"}--end--

JSON 字段要求：
- overall：本群讨论风格的整体评价，例如活跃、太水、话题不集中、技术氛围浓、娱乐性强等，50 到 160 字。
- topics：不多于 10 个重点话题，按热度或重要性排序。
- topics[].title：话题名，50 字以内，不要带序号，不要带火焰符号。
- topics[].heat：热度，1 到 5 的整数，数字越大代表越热。
- topics[].participants：参与者，不超过 5 人，去重。
- topics[].start_time 和 topics[].end_time：时间段，格式 HH:MM。
- topics[].process：话题过程，50 到 200 字。
- topics[].comment：简短评价，50 字以下。
- resources：群友分享的链接资源；没有资源时返回空数组。resources[].title 是资源名称，resources[].url 是链接，resources[].description 是简短说明。

总结结果需要适合同时渲染为文本消息和图片海报。输出必须是合法 JSON。`

	msg := fmt.Sprintf("群名称: %s\n聊天记录如下:\n%s", chatRoomName, transcript)
	return []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(prompt),
		openai.UserMessage(msg),
	}
}

func requestChatRoomSummaryReport(ctx context.Context, client *openai.Client, summaryModel string, messages []openai.ChatCompletionMessageParamUnion, withSchema bool) (*chatRoomSummaryReport, error) {
	req := openai.ChatCompletionNewParams{
		Model:               summaryModel,
		Messages:            messages,
		MaxCompletionTokens: openai.Int(chatRoomSummaryMaxCompletionTokens),
	}
	if withSchema {
		req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        "chat_room_summary_report",
					Description: openai.String("微信群聊总结报告，用于同时渲染文本总结和图片海报。"),
					Strict:      openai.Bool(false),
					Schema:      chatRoomSummarySchema(),
				},
			},
		}
	}

	msg, err := streamChatCompletionMessage(ctx, client, req)
	if err != nil {
		return nil, err
	}

	var report chatRoomSummaryReport
	if err := json.Unmarshal([]byte(cleanJSONContent(msg.Content)), &report); err != nil {
		return nil, fmt.Errorf("解析群聊总结 JSON 失败: %w", err)
	}
	return normalizeChatRoomSummaryReport(&report), nil
}

func chatRoomSummarySchema() map[string]any {
	topic := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title":        map[string]any{"type": "string", "description": "话题名，50 字以内，不带序号和火焰符号。"},
			"heat":         map[string]any{"type": "integer", "minimum": 1, "maximum": 5},
			"participants": map[string]any{"type": "array", "maxItems": 5, "items": map[string]any{"type": "string"}},
			"start_time":   map[string]any{"type": "string", "description": "开始时间，格式 HH:MM。"},
			"end_time":     map[string]any{"type": "string", "description": "结束时间，格式 HH:MM。"},
			"process":      map[string]any{"type": "string", "description": "话题过程，50 到 200 字。"},
			"comment":      map[string]any{"type": "string", "description": "简短评价，50 字以下。"},
		},
		"required":             []string{"title", "heat", "participants", "start_time", "end_time", "process", "comment"},
		"additionalProperties": false,
	}
	resource := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title":       map[string]any{"type": "string"},
			"url":         map[string]any{"type": "string"},
			"description": map[string]any{"type": "string"},
		},
		"required":             []string{"title", "url", "description"},
		"additionalProperties": false,
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"overall":   map[string]any{"type": "string"},
			"topics":    map[string]any{"type": "array", "maxItems": 10, "items": topic},
			"resources": map[string]any{"type": "array", "items": resource},
		},
		"required":             []string{"overall", "topics", "resources"},
		"additionalProperties": false,
	}
}

func normalizeChatRoomSummaryReport(report *chatRoomSummaryReport) *chatRoomSummaryReport {
	report.Overall = strings.TrimSpace(report.Overall)
	if len(report.Topics) > 10 {
		report.Topics = report.Topics[:10]
	}
	for idx := range report.Topics {
		topic := &report.Topics[idx]
		topic.Title = strings.TrimSpace(topic.Title)
		topic.Process = strings.TrimSpace(topic.Process)
		topic.Comment = strings.TrimSpace(topic.Comment)
		topic.StartTime = strings.TrimSpace(topic.StartTime)
		topic.EndTime = strings.TrimSpace(topic.EndTime)
		if topic.Heat < 1 {
			topic.Heat = 1
		}
		if topic.Heat > 5 {
			topic.Heat = 5
		}
		if len(topic.Participants) > 5 {
			topic.Participants = topic.Participants[:5]
		}
		seen := make(map[string]struct{}, len(topic.Participants))
		participants := make([]string, 0, len(topic.Participants))
		for _, participant := range topic.Participants {
			participant = strings.TrimSpace(participant)
			if participant == "" {
				continue
			}
			if _, ok := seen[participant]; ok {
				continue
			}
			seen[participant] = struct{}{}
			participants = append(participants, participant)
		}
		topic.Participants = participants
	}

	resources := make([]chatRoomSummaryResource, 0, len(report.Resources))
	for _, resource := range report.Resources {
		resource.Title = strings.TrimSpace(resource.Title)
		resource.URL = strings.TrimSpace(resource.URL)
		resource.Description = strings.TrimSpace(resource.Description)
		if resource.Title == "" && resource.URL == "" && resource.Description == "" {
			continue
		}
		resources = append(resources, resource)
	}
	report.Resources = resources
	return report
}

func renderChatRoomSummaryText(summaryModel string, report *chatRoomSummaryReport) string {
	var builder strings.Builder
	builder.WriteString("#消息总结\n")
	builder.WriteString("让我们一起来看看群友们都聊了什么有趣的话题吧~\n\n")
	builder.WriteString(fmt.Sprintf("本次总结由**%s**加持\n\n", summaryModel))
	builder.WriteString(fmt.Sprintf("整体评价：%s\n\n", report.Overall))

	for idx, topic := range report.Topics {
		title := fmt.Sprintf("%s %s %s", numberEmoji(idx+1), topic.Title, strings.Repeat("🔥", topic.Heat))
		builder.WriteString(title)
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("- 话题名：%s\n", title))
		builder.WriteString(fmt.Sprintf("- 参与者：%s\n", strings.Join(topic.Participants, "、")))
		builder.WriteString(fmt.Sprintf("- 时间段：%s 至 %s\n", topic.StartTime, topic.EndTime))
		builder.WriteString(fmt.Sprintf("- 过程：%s\n", topic.Process))
		builder.WriteString(fmt.Sprintf("- 评价：%s\n", topic.Comment))
		builder.WriteString("------------\n")
	}

	if len(report.Resources) > 0 {
		builder.WriteString("\n群友分享链接资源：\n")
		for _, resource := range report.Resources {
			resourceText := resource.Title
			if resource.URL != "" {
				resourceText += "：" + resource.URL
			}
			if resource.Description != "" {
				resourceText += "（" + resource.Description + "）"
			}
			builder.WriteString("- ")
			builder.WriteString(resourceText)
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func buildChatRoomSummaryTemplateData(chatRoomName, summaryModel string, report *chatRoomSummaryReport) chatRoomSummaryTemplateData {
	topics := make([]chatRoomSummaryTopicView, 0, len(report.Topics))
	maxHeat := 0
	for idx, topic := range report.Topics {
		if topic.Heat > maxHeat {
			maxHeat = topic.Heat
		}
		topics = append(topics, chatRoomSummaryTopicView{
			Number:       numberEmoji(idx + 1),
			Title:        topic.Title,
			HeatIcon:     strings.Repeat("🔥", topic.Heat),
			HeatLabel:    heatLabel(topic.Heat),
			TimeRange:    fmt.Sprintf("%s - %s", topic.StartTime, topic.EndTime),
			Participants: topic.Participants,
			Process:      topic.Process,
			Comment:      topic.Comment,
			Accent:       topicAccent(topic.Heat, idx),
			SpanClass:    topicSpanClass(idx, len(report.Topics)),
		})
	}

	resources := make([]chatRoomSummaryResourceView, 0, len(report.Resources))
	for _, resource := range report.Resources {
		resources = append(resources, chatRoomSummaryResourceView{
			Title:       defaultString(resource.Title, "群友分享资源"),
			URL:         resource.URL,
			Description: resource.Description,
			Icon:        resourceIcon(resource),
		})
	}
	maxHeatText := strings.Repeat("🔥", maxHeat)
	if maxHeatText == "" {
		maxHeatText = "0"
	}

	return chatRoomSummaryTemplateData{
		Title:         fmt.Sprintf("%s早报", chatRoomName),
		SummaryModel:  summaryModel,
		ChatRoomName:  chatRoomName,
		GeneratedAt:   time.Now().Format("2006-01-02 15:04"),
		Overall:       report.Overall,
		TopicCount:    len(report.Topics),
		ResourceCount: len(report.Resources),
		MaxHeatText:   maxHeatText,
		Topics:        topics,
		Resources:     resources,
	}
}

func renderChatRoomSummaryHTML(data chatRoomSummaryTemplateData) (string, error) {
	templateBytes, err := chatRoomSummaryTemplate.FS.ReadFile("chat_room_summary.html")
	if err != nil {
		return "", err
	}
	tpl, err := htmltemplate.New("chat_room_summary.html").Parse(string(templateBytes))
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (s *ChatRoomService) sendChatRoomSummaryImage(ctx context.Context, msgService *MessageService, chatRoomID string, data chatRoomSummaryTemplateData) error {
	htmlContent, err := renderChatRoomSummaryHTML(data)
	if err != nil {
		return fmt.Errorf("渲染群聊总结模板失败: %w", err)
	}
	pngBytes, err := captureHTMLScreenshot(ctx, htmlContent)
	if err != nil {
		return fmt.Errorf("群聊总结截图失败: %w", err)
	}
	_, err = msgService.MsgUploadImg(chatRoomID, bytes.NewReader(pngBytes))
	if err != nil {
		return fmt.Errorf("发送群聊总结图片失败: %w", err)
	}
	return nil
}

func captureHTMLScreenshot(ctx context.Context, htmlContent string) ([]byte, error) {
	tempFile, err := os.CreateTemp("", "chat_room_summary_*.html")
	if err != nil {
		return nil, err
	}
	tempFileName := tempFile.Name()
	defer os.Remove(tempFileName)

	if _, err := tempFile.WriteString(htmlContent); err != nil {
		tempFile.Close()
		return nil, err
	}
	if err := tempFile.Close(); err != nil {
		return nil, err
	}

	fileURL := url.URL{Scheme: "file", Path: tempFileName}
	allocatorOptions := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(960, 1000), // 减小窗口初始宽度，适应手机屏幕阅读
	)
	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(ctx, allocatorOptions...)
	defer allocatorCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocatorCtx)
	defer browserCancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(browserCtx, 45*time.Second)
	defer timeoutCancel()

	var pngBytes []byte
	if err := chromedp.Run(timeoutCtx,
		chromedp.EmulateViewport(960, 0, chromedp.EmulateScale(2)), // 宽度 960 2 开启视网膜高清分辨率
		chromedp.Navigate(fileURL.String()),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Evaluate(`document.fonts ? document.fonts.ready.then(() => true) : true`, nil),
		chromedp.FullScreenshot(&pngBytes, 100), // FullScreenshot 会自动计算页面的实际高度进行全尺寸截图
	); err != nil {
		return nil, err
	}
	return pngBytes, nil
}

func numberEmoji(number int) string {
	switch number {
	case 1:
		return "1️⃣"
	case 2:
		return "2️⃣"
	case 3:
		return "3️⃣"
	case 4:
		return "4️⃣"
	case 5:
		return "5️⃣"
	case 6:
		return "6️⃣"
	case 7:
		return "7️⃣"
	case 8:
		return "8️⃣"
	case 9:
		return "9️⃣"
	case 10:
		return "🔟"
	default:
		return fmt.Sprintf("%d.", number)
	}
}

func heatLabel(heat int) string {
	switch {
	case heat >= 5:
		return "热度 MAX"
	case heat >= 3:
		return "极高热度"
	case heat == 2:
		return "高热度"
	default:
		return "普通热度"
	}
}

func topicAccent(heat, index int) string {
	if heat >= 5 {
		return "rose"
	}
	if heat >= 3 {
		return "orange"
	}
	accents := []string{"cyan", "purple", "green", "yellow"}
	return accents[index%len(accents)]
}

func topicSpanClass(index, total int) string {
	if total > 3 && total%3 == 1 && index == total-1 {
		return "span-3"
	}
	return ""
}

func resourceIcon(resource chatRoomSummaryResource) string {
	lowerText := strings.ToLower(resource.Title + " " + resource.URL + " " + resource.Description)
	switch {
	case strings.Contains(lowerText, "github"):
		return "GH"
	case strings.Contains(lowerText, "douyin") || strings.Contains(lowerText, "抖音") || strings.Contains(lowerText, "tiktok"):
		return "♪"
	case strings.Contains(lowerText, "新闻") || strings.Contains(lowerText, "news"):
		return "N"
	default:
		return "↗"
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
