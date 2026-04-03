package plugins

import (
	"testing"

	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
)

func TestDouYinVideoParse(t *testing.T) {
	parser := NewDouyinVideoParsePlugin()
	parser.Run(&plugin.MessageContext{
		Message: &model.Message{
			// 1. 0.76 复制打开抖音，看看【军 哥的作品】# 沿途风景随拍蓝天白云 # 再好的副驾驶不如自己... https://v.douyin.com/UHLvzH3alyM/ 06/05 u@f.bA WMW:/
			// 2. 0.53 蹬转教学基础版 # 羽毛球 # 羽毛球培训 # 教练我想练球  https://v.douyin.com/mYgqYsYw_RA/ 复制此链接，打开抖音搜索，直接观看视频！ nqR:/ 10/12 P@k.Ch
			Content: `0.76 复制打开抖音，看看【军 哥的作品】# 沿途风景随拍蓝天白云 # 再好的副驾驶不如自己... https://v.douyin.com/UHLvzH3alyM/ 06/05 u@f.bA WMW:/`,
		},
	})
}
