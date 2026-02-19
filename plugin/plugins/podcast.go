package plugins

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/plugin/pkg/podcast"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type PodcastPlugin struct {
	PodcastSecrets podcast.PodcastSecrets
	PodcastConfig  podcast.PodcastConfig
}

func NewPodcastPlugin() plugin.MessageHandler {
	return &PodcastPlugin{}
}

func (p *PodcastPlugin) GetName() string {
	return "Podcast"
}

func (p *PodcastPlugin) GetLabels() []string {
	return []string{"text", "chat"}
}

func (p *PodcastPlugin) Match(ctx *plugin.MessageContext) bool {
	return strings.HasPrefix(ctx.MessageContent, "#AI播客")
}

func (p *PodcastPlugin) PreAction(ctx *plugin.MessageContext) bool {
	if !NewChatRoomCommonPlugin().PreAction(ctx) {
		return false
	}

	chatRoomSettings, err := service.NewChatRoomSettingsService(context.Background()).GetChatRoomSettings(ctx.Message.FromWxID)
	if err != nil || chatRoomSettings == nil {
		return false
	}
	if chatRoomSettings.PodcastEnabled == nil || !*chatRoomSettings.PodcastEnabled {
		return false
	}

	type providerSecret struct {
		AppID      string `json:"app_id"`
		AccessKey  string `json:"access_key"`
		ResourceID string `json:"resource_id"`
	}

	providerSecrets := make(map[string]providerSecret)
	if len(chatRoomSettings.PodcastConfig) == 0 {
		return false
	}
	if err := json.Unmarshal(chatRoomSettings.PodcastConfig, &providerSecrets); err != nil {
		return false
	}

	doubaoSecret, ok := providerSecrets["DouBao"]
	if !ok {
		return false
	}

	if doubaoSecret.AppID == "" || doubaoSecret.AccessKey == "" {
		return false
	}

	p.PodcastSecrets = podcast.PodcastSecrets{
		AppID:       doubaoSecret.AppID,
		AccessToken: doubaoSecret.AccessKey,
		ResourceID:  doubaoSecret.ResourceID,
	}

	return true
}

func (p *PodcastPlugin) PostAction(ctx *plugin.MessageContext) {
}

func (p *PodcastPlugin) Run(ctx *plugin.MessageContext) {
	if !p.PreAction(ctx) {
		return
	}

	if ctx.Message.Type == model.MsgTypeText {
		if err := p.InitPodcastConfigFromTextMessage(ctx.MessageContent); err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return
		}
	} else if ctx.ReferMessage != nil {
		switch ctx.ReferMessage.Type {
		case model.MsgTypeText:
			if err := p.InitPodcastConfigFromTextMessage(ctx.ReferMessage.Content); err != nil {
				ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
				return
			}
		case model.MsgTypeApp:
			if ctx.ReferMessage.AppMsgType == model.AppMsgTypeUrl {
				var xmlMessage robot.XmlMessage
				if err := vars.RobotRuntime.XmlDecoder(ctx.ReferMessage.Content, &xmlMessage); err != nil {
					ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "引用消息解析失败")
					return
				}
				p.PodcastConfig.Action = 0
				p.PodcastConfig.InputInfo.InputURL = strings.ReplaceAll(xmlMessage.AppMsg.URL, "&amp;", "&")
			} else {
				ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "暂不支持的消息类型")
				return
			}
		default:
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "暂不支持的消息类型")
			return
		}
	}

	var audioURL string
	var err error

	audioURL, err = podcast.Podcast(p.PodcastSecrets, p.PodcastConfig)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return
	}

	var podcastMessage robot.XmlMessage
	podcastMessage.AppMsg = robot.AppMessage{
		AppID:        "wx281a70a3d390bdf2",
		SDKVer:       "0",
		Action:       "view",
		Type:         76,
		ShowType:     0,
		SoundType:    0,
		Title:        "AI 播客",
		Des:          "吾有一言，请诸君静听",
		MediaTagName: "播客",
		URL:          "https://i.y.qq.com/v8/playsong.html?hosteuin=Nenqow65oiCk&sharefrom=&from_id=59110203&from_idtype=10025&from_name=JUU3JTgzJUFEJUU3JTgyJUI5JUU2JTkyJUFEJUU1JUFFJUEy&songid=563780836&songmid=&type=0&platform=(10rpl)&appsongtype=(11rpl)&_wv=1&source=qq&appshare=iphone&media_mid=000GUagT0MejGR&ADTAG=wxfshare",
		DataURL:      audioURL,
		SongAlbumURL: "http://wxapp.tc.qq.com/202/20304/stodownload?filekey=30350201010421301f020200ca0402535a04103795db33b0f79bf455ac175510a03f190203019df9040d00000004627466730000000132&hy=SZ&storeid=26996ff3e000997b45696895c000000ca00004f50535a14a908e0b6bbd461f&bizid=1023",
		AppAttach: robot.AppAttach{
			CDNThumbURL:    "3057020100044b304902010002045696895c02032f5bd502046c621f7402046996ff41042432666266383235632d653161642d346231322d613033612d6264663031356361323361370204012808030201000405004c57c100",
			CDNThumbMD5:    "08621737cb1bcb15e17cf799633153dd",
			CDNThumbLength: 24859,
			CDNThumbWidth:  450,
			CDNThumbHeight: 450,
			CDNThumbAESKey: "77ccaf72f6715709e196983206f551d1",
			AESKey:         "77ccaf72f6715709e196983206f551d1",
			EncryVer:       0,
		},
		SourceDisplayName: "播客",
		ContentAttr:       0,
		StreamVideo:       0,
		StatExtStr:        "GhQKEnd4NWFhMzMzNjA2NTUwZGZkNQ==",
	}

	appMessageBytes, err := xml.Marshal(podcastMessage.AppMsg)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "播客消息生成失败")
		return
	}

	err = ctx.MessageService.SendAppMessage(ctx.Message.FromWxID, 76, string(appMessageBytes))
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "播客消息发送失败")
		return
	}
}

func (p *PodcastPlugin) InitPodcastConfigFromTextMessage(message string) error {
	podcastContent := strings.TrimPrefix(message, "#AI播客")
	podcastContent = strings.TrimSpace(podcastContent)
	if podcastContent == "" {
		return fmt.Errorf("请输入播客内容，格式: #AI播客 <文本内容或音频链接>")
	}

	if strings.HasPrefix(podcastContent, "http") {
		p.PodcastConfig.Action = 0
		p.PodcastConfig.InputInfo.InputURL = podcastContent

	} else {
		p.PodcastConfig.Action = 4
		p.PodcastConfig.PromptText = podcastContent
	}
	return nil
}
