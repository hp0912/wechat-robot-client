package utils

import (
	"regexp"
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
