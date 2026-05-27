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
			// 3. 1.00 复制打开抖音，看看【陈老实的作品】恋爱脑把视频看十遍 # 陈老实 # 恋爱军师陈哥  https://v.douyin.com/iywNyE8DHEQ/ odN:/ h@b.NW 11/21
			Content: `1.00 复制打开抖音，看看【陈老实的作品】恋爱脑把视频看十遍 # 陈老实 # 恋爱军师陈哥  https://v.douyin.com/iywNyE8DHEQ/ odN:/ h@b.NW 11/21`,
		},
	})
}
