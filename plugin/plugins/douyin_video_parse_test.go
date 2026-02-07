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
			Content: `0.53 蹬转教学基础版 # 羽毛球 # 羽毛球培训 # 教练我想练球  https://v.douyin.com/mYgqYsYw_RA/ 复制此链接，打开抖音搜索，直接观看视频！ nqR:/ 10/12 P@k.Ch `,
		},
	})
}
