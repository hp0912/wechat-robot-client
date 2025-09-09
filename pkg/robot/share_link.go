package robot

import "encoding/xml"

type ShareLinkMessage struct {
	XMLName     xml.Name `xml:"appmsg"`
	AppID       string   `xml:"appid,attr"`
	SDKVer      string   `xml:"sdkver,attr"`
	Title       string   `xml:"title"`
	Des         string   `xml:"des"`
	Type        int      `xml:"type"`
	ShowType    int      `xml:"showtype"`
	SoundType   int      `xml:"soundtype"`
	ContentAttr int      `xml:"contentattr"`
	DirectShare int      `xml:"directshare"`
	Url         string   `xml:"url"`
	ThumbUrl    string   `xml:"thumburl"`
	AppAttach   string   `xml:"appattach"`
}
