package service

import (
	"context"
	"os"
	"testing"
)

func Test_captureHTMLScreenshot(t *testing.T) {
	data := chatRoomSummaryTemplateData{
		Title:         "测试群早报",
		SummaryModel:  "qwen3.6-plus",
		ChatRoomName:  "WeChat Robot Ultra",
		GeneratedAt:   "2026-05-04 07:31",
		Overall:       "本群技术氛围浓厚且极具极客精神，日常围绕AI大模型应用、微信机器人开发、自动化脚本及账号共享展开深度探讨。群友互动频繁，既有硬核的代码部署与MCP架构交流，也不乏签到修仙、新闻播报等趣味玩法，整体呈现出高效务实与轻松娱乐并存的活跃生态。",
		TopicCount:    6,
		ResourceCount: 3,
		MaxHeatText:   "🔥🔥🔥🔥🔥",
		Topics: []chatRoomSummaryTopicView{
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
			{
				Number:       "1️⃣",
				Title:        "GPT Plus土耳其区订阅与账号共享",
				HeatIcon:     "🔥🔥🔥🔥🔥",
				HeatLabel:    "热度 MAX",
				TimeRange:    "11:02 - 11:13",
				Participants: []string{"光光", "水牛", "KAITO", "♥️", "听夜雨 出红尘皆是"},
				Process:      "群友分享通过土耳其区App Store购买GPT Plus的经验，对比官方价格与代买差价，探讨使用咸鱼礼品卡充值及账号共享的可行性，并交流了注册外区账号遇到的验证门槛与解决思路。",
				Comment:      "实用省钱攻略，外区订阅成主流选择。",
				Accent:       "rose",
				SpanClass:    "",
			},
		},
		Resources: []chatRoomSummaryResourceView{},
	}

	htmlContent, err := renderChatRoomSummaryHTML(data)
	if err != nil {
		t.Fatalf("renderChatRoomSummaryHTML failed: %v", err)
	}

	pngBytes, err := captureHTMLScreenshot(context.Background(), htmlContent)
	if err != nil {
		t.Fatalf("captureHTMLScreenshot failed: %v", err)
	}

	err = os.WriteFile("test_summary.png", pngBytes, 0644)
	if err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	t.Logf("Successfully captured screenshot, saved to test_summary.png")
}
