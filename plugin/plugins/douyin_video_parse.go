package plugins

import (
	"fmt"
	"net/http"
	"regexp"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/plugin/pkg"

	"github.com/go-resty/resty/v2"
)

type VideoParseResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Author        string   `json:"author"`
		AuthorID      string   `json:"author_id"`
		Like          int32    `json:"like"`
		Comment       int32    `json:"comment"`
		Collect       int32    `json:"collect"`
		Share         int32    `json:"share"`
		CreateTime    int64    `json:"create_time"`
		Title         string   `json:"title"`
		Cover         string   `json:"cover"`
		URL           string   `json:"url"`
		VideoDuration int32    `json:"video_duration"`
		MusicURL      string   `json:"music_url"`
		MusicCover    string   `json:"music_cover"`
		Images        []string `json:"images"`
	} `json:"data"`
}

type DouyinVideoParsePlugin struct{}

func NewDouyinVideoParsePlugin() plugin.MessageHandler {
	return &DouyinVideoParsePlugin{}
}

func (p *DouyinVideoParsePlugin) GetName() string {
	return "DouyinVideoParse"
}

func (p *DouyinVideoParsePlugin) GetLabels() []string {
	return []string{"text", "douyin"}
}

func (p *DouyinVideoParsePlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *DouyinVideoParsePlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *DouyinVideoParsePlugin) Run(ctx *plugin.MessageContext) bool {
	var douyinShareContent string
	if ctx.ReferMessage != nil {
		douyinShareContent = ctx.ReferMessage.Content
	} else {
		douyinShareContent = ctx.Message.Content
	}

	re := regexp.MustCompile(`https://[^\s]+`)
	matches := re.FindAllString(douyinShareContent, -1)
	if len(matches) == 0 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "未找到抖音链接")
		return true
	}

	// 获取第一个匹配的链接
	douyinURL := matches[0]

	var respData VideoParseResponse
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("url", douyinURL).
		SetResult(&respData).
		Get("https://api.pearktrue.cn/api/video/douyin/")
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return true
	}
	if resp.StatusCode() != http.StatusOK {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, http.StatusText(resp.StatusCode()))
		return true
	}
	if respData.Data.URL == "" {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "解析失败，可能是链接已失效或格式不正确")
		return true
	}

	_ = ctx.MessageService.ShareLink(ctx.Message.FromWxID, robot.ShareLinkMessage{
		Title:    fmt.Sprintf("抖音视频解析成功 - %s", respData.Data.Author),
		Des:      respData.Data.Title,
		Url:      respData.Data.URL,
		ThumbUrl: robot.CDATAString("https://mmbiz.qpic.cn/mmbiz_png/NbW0ZIUM8lVHoUbjXw2YbYXbNJDtUH7Sbkibm9Qwo9FhAiaEFG4jY3Q2MEleRpiaWDyDv8BZUfR85AW3kG4ib6DyAw/640?wx_fmt=png"),
	})

	_ = pkg.SendVideoByURL(ctx.MessageService, ctx.Message.FromWxID, respData.Data.URL)

	return true
}
