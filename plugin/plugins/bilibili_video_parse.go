package plugins

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"

	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/vars"
)

// BilibiliAPIResponse 第三方接口返回的 B站解析结果
type BilibiliAPIResponse struct {
	Code      int             `json:"code"`
	Msg       string          `json:"msg"`
	Data      BilibiliAPIData `json:"data"`
	APISource string          `json:"api_source"`
}

// BilibiliAPIData B站视频解析数据
type BilibiliAPIData struct {
	InputURL  string `json:"input_url"`
	Author    string `json:"author"`
	Avatar    string `json:"avatar"`
	View      int64  `json:"view"`
	Danmu     int64  `json:"danmu"`
	Reply     int64  `json:"reply"`
	Coin      int64  `json:"coin"`
	Favorite  int64  `json:"favorite"`
	Share     int64  `json:"share"`
	Like      int64  `json:"like"`
	Pubdate   string `json:"pubdate"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	Cover     string `json:"cover"`
	URL       string `json:"url"`
	VideoSize string `json:"video_size"`
}

type BilibiliVideoParsePlugin struct{}

func NewBilibiliVideoParsePlugin() plugin.MessageHandler {
	return &BilibiliVideoParsePlugin{}
}

func (p *BilibiliVideoParsePlugin) GetName() string {
	return "BilibiliVideoParse"
}

func (p *BilibiliVideoParsePlugin) GetLabels() []string {
	return []string{"text", "bilibili"}
}

func (p *BilibiliVideoParsePlugin) PreAction(ctx *plugin.MessageContext) bool {
	if ctx.Message.IsChatRoom {
		next := NewChatRoomCommonPlugin().PreAction(ctx)
		if !next {
			return false
		}
		if !ctx.Settings.IsShortVideoParsingEnabled() {
			return false
		}
	}
	return true
}

func (p *BilibiliVideoParsePlugin) PostAction(ctx *plugin.MessageContext) {}

func (p *BilibiliVideoParsePlugin) Match(ctx *plugin.MessageContext) bool {
	if ctx.ReferMessage != nil {
		return false
	}
	return strings.Contains(ctx.Message.Content, "https://www.bilibili.com/video")
}

func (p *BilibiliVideoParsePlugin) Run(ctx *plugin.MessageContext) {
	if !p.PreAction(ctx) {
		return
	}

	re := regexp.MustCompile(`https://[^\s]+`)
	matches := re.FindAllString(ctx.Message.Content, -1)
	if len(matches) == 0 {
		return
	}
	bilibiliURL := matches[0]

	respData, err := p.ParseBilibiliVideo(bilibiliURL)
	if err != nil {
		log.Printf("Bilibili视频解析失败: %v\n", err)
		return
	}

	if respData.Data.URL == "" {
		log.Printf("Bilibili视频解析成功但未获取到分享链接\n")
		return
	}

	shareLink := robot.ShareLinkMessage{
		Title:    fmt.Sprintf("B站视频 - %s", respData.Data.Author),
		Des:      respData.Data.Title,
		Url:      respData.Data.URL,
		ThumbUrl: robot.CDATAString("https://mmbiz.qpic.cn/sz_mmbiz_jpg/lB5IHibX4CX3ibibHThIgqecGpt0Xv98fkia1UcgIqiaEnZCDnibhY5qb7CZytpwQ6F9zSLo37ricYz8bfOEMuiclozTvGxQLfHsJia5LKNEk2Cpekp8/640?wx_fmt=jpeg&amp;from=appmsg"),
	}
	if respData.Data.Desc != "" {
		shareLink.Des = respData.Data.Desc
	}

	err = ctx.MessageService.ShareLink(ctx.Message.FromWxID, shareLink)
	if err != nil {
		log.Printf("发送Bilibili分享链接失败: %v\n", err)
	}
}

func (p *BilibiliVideoParsePlugin) ParseBilibiliVideo(bilibiliURL string) (BilibiliAPIResponse, error) {
	var respData BilibiliAPIResponse
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"key": vars.ThirdPartyApiKey,
			"url": bilibiliURL,
		}).
		SetResult(&respData).
		Post("https://api.pearapi.ai/api/video/api.php")
	if err != nil {
		return BilibiliAPIResponse{}, err
	}
	if resp.StatusCode() != http.StatusOK {
		return BilibiliAPIResponse{}, fmt.Errorf("外部接口请求失败，状态码: %d %s", resp.StatusCode(), http.StatusText(resp.StatusCode()))
	}
	if respData.Code != http.StatusOK {
		return BilibiliAPIResponse{}, fmt.Errorf("外部接口解析失败，code: %d, msg: %s", respData.Code, respData.Msg)
	}
	return respData, nil
}
