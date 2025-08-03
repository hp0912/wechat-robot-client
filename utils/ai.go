package utils

import (
	"regexp"
	"strings"
	"wechat-robot-client/vars"
)

func TrimAt(content string) string {
	// 去除@开头的触发词
	re := regexp.MustCompile(vars.TrimAtRegexp)
	return re.ReplaceAllString(content, "")
}

func TrimAITriggerWord(content, aiTriggerWord string) string {
	// 去除固定AI触发词
	re := regexp.MustCompile("^" + regexp.QuoteMeta(aiTriggerWord) + `[\s，,：:]*`)
	return re.ReplaceAllString(content, "")
}

func TrimAITriggerAll(content, aiTriggerWord string) string {
	return TrimAITriggerWord(TrimAt(content), aiTriggerWord)
}

// NormalizeAIBaseURL 规范化AI BaseURL，确保以/v+数字结尾，如果没有则添加/v1
func NormalizeAIBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	versionRegex := regexp.MustCompile(`/v\d+$`)
	if !versionRegex.MatchString(baseURL) {
		baseURL += "/v1"
	}
	return baseURL
}
