package service

import (
	"strings"
	"testing"

	"wechat-robot-client/model"
)

func TestBuildUnderstandVideoRequestUsesTextWithVideoURL(t *testing.T) {
	req, err := buildUnderstandVideoRequest("  https://example.com/video.mp4  ", model.MomentSettings{
		VideoUnderstandingModel: "video-model",
	})
	if err != nil {
		t.Fatalf("buildUnderstandVideoRequest returned error: %v", err)
	}
	if req.Model != "video-model" {
		t.Fatalf("model = %q, want %q", req.Model, "video-model")
	}
	if len(req.Messages) != 2 {
		t.Fatalf("message count = %d, want 2", len(req.Messages))
	}

	content, ok := req.Messages[1].GetContent().AsAny().(*string)
	if !ok {
		t.Fatalf("user message content type = %T, want *string", req.Messages[1].GetContent().AsAny())
	}
	if !strings.Contains(*content, "请理解这个朋友圈视频内容") {
		t.Fatalf("user message content does not contain prompt: %q", *content)
	}
	if !strings.Contains(*content, "视频链接：https://example.com/video.mp4") {
		t.Fatalf("user message content does not contain trimmed video URL: %q", *content)
	}
}
