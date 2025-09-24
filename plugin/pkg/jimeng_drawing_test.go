package pkg

import "testing"

func TestJimengDrawing(t *testing.T) {
	urls, err := JimengDrawing(&JimengConfig{
		BaseURL:   "http://127.0.0.1:9000",
		SessionID: []string{"6d31ec4badd96cb83b4ac9e90a179a7c"},
		JimengRequest: JimengRequest{
			Model:          "jimeng-4.0",
			Prompt:         "圣诞写真，中国美女，黑色抹胸圣诞裙，黑白圣诞帽，纯欲风格，有银色蝴蝶结的圣诞树，清透有质感，全身照，圣诞氛围感，有欧式室内家居，闪光灯效果，手拿银光星星，欧式火炉，微笑",
			NegativePrompt: "",
			Width:          1024,
			Height:         1024,
			SampleStrength: 0.5,
		},
	})
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(urls)
}
