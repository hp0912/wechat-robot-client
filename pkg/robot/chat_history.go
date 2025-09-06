package robot

import "encoding/xml"

type ChatHistoryMessage struct {
	XMLName      xml.Name           `xml:"msg"`
	AppMsg       ChatHistoryAppMsg  `xml:"appmsg"`
	FromUsername string             `xml:"fromusername,omitempty"`
	Scene        int                `xml:"scene,omitempty"`
	AppInfo      ChatHistoryAppInfo `xml:"appinfo,omitempty"`
	CommentURL   string             `xml:"commenturl,omitempty"`
}

type ChatHistoryAppMsg struct {
	XMLName    xml.Name              `xml:"appmsg"`
	AppID      string                `xml:"appid,attr"`
	SDKVer     string                `xml:"sdkver,attr"`
	Title      string                `xml:"title"`
	Des        string                `xml:"des"`
	Type       int                   `xml:"type"`
	URL        string                `xml:"url"`
	AppAttach  ChatHistoryAppAttach  `xml:"appattach"`
	RecordItem ChatHistoryRecordItem `xml:"recorditem"`
	Percent    int                   `xml:"percent,omitempty"`
}

type ChatHistoryRecordItem struct {
	XML string `xml:",innerxml"`
}

type ChatHistoryAppAttach struct {
	CDNThumbAESKey string `xml:"cdnthumbaeskey"`
	AESKey         string `xml:"aeskey"`
}

type ChatHistoryAppInfo struct {
	Version int    `xml:"version"`
	AppName string `xml:"appname"`
}

type RecordInfo struct {
	XMLName    xml.Name `xml:"recordinfo"`
	Info       string   `xml:"info"`
	IsChatRoom int      `xml:"isChatRoom"`
	DataList   DataList `xml:"datalist"`
	Desc       string   `xml:"desc"`
	FromScene  int      `xml:"fromscene"`
}

type DataList struct {
	Count int        `xml:"count,attr"`
	Items []DataItem `xml:"dataitem"`
}

type DataItem struct {
	DataType         int                   `xml:"datatype,attr"`
	DataID           string                `xml:"dataid,attr"`
	MessageUUID      string                `xml:"messageuuid,omitempty"`
	SrcMsgLocalID    int                   `xml:"srcMsgLocalid"`
	SourceTime       string                `xml:"sourcetime"`
	FromNewMsgID     int64                 `xml:"fromnewmsgid"`
	SrcMsgCreateTime int64                 `xml:"srcMsgCreateTime"`
	SourceName       string                `xml:"sourcename"`
	SourceHeadURL    string                `xml:"sourceheadurl"`
	DataTitle        string                `xml:"datatitle,omitempty"`
	DataDesc         string                `xml:"datadesc"`
	IsChatRoom       int                   `xml:"ischatroom,omitempty"`
	RecordXML        *RecordXML            `xml:"recordxml,omitempty"`
	EmojiItem        *ChatHistoryEmojiItem `xml:"emojiitem,omitempty"`
	FinderFeed       *FinderFeed           `xml:"finderFeed,omitempty"`
	WebURLItem       *WebURLItem           `xml:"weburlitem,omitempty"`
	DataItemSource   *DataItemSource       `xml:"dataitemsource"`
	ThumbSize        int                   `xml:"thumbsize,omitempty"`
	CDNThumbURL      string                `xml:"cdnthumburl,omitempty"`
	ThumbFileType    int                   `xml:"thumbfiletype,omitempty"`
	DataFmt          string                `xml:"datafmt,omitempty"`
	ThumbHeight      int                   `xml:"thumbheight,omitempty"`
	CDNDataKey       string                `xml:"cdndatakey,omitempty"`
	DataSize         int                   `xml:"datasize,omitempty"`
	ThumbFullMD5     string                `xml:"thumbfullmd5,omitempty"`
	FileType         int                   `xml:"filetype,omitempty"`
	CDNThumbKey      string                `xml:"cdnthumbkey,omitempty"`
	ThumbWidth       int                   `xml:"thumbwidth,omitempty"`
	CDNDataURL       string                `xml:"cdndataurl,omitempty"`
	FullMD5          string                `xml:"fullmd5,omitempty"`
	Link             string                `xml:"link,omitempty"`
}

type RecordXML struct {
	RecordInfo RecordInfo `xml:"recordinfo"`
}

type DataItemSource struct {
	HashUsername string `xml:"hashusername"`
}

type ChatHistoryEmojiItem struct {
	CDNURLString     string `xml:"cdnurlstring"`
	ExternURL        string `xml:"externurl"`
	UIEmoticonWidth  int    `xml:"uiemoticonwidth"`
	ExternMD5        string `xml:"externmd5"`
	EncryptURLString string `xml:"encrypturlstring"`
	UIEmoticonHeight int    `xml:"uiemoticonheight"`
	MD5              string `xml:"md5"`
	UIEmoticonType   int    `xml:"uiemoticontype"`
}

type WebURLItem struct {
	ThumbURL string `xml:"thumburl"`
	Title    string `xml:"title"`
	Link     string `xml:"link"`
	Desc     string `xml:"desc"`
}

type FinderFeed struct {
	BizAuthIconType    int              `xml:"bizAuthIconType"`
	FeedType           int              `xml:"feedType"`
	Nickname           string           `xml:"nickname"`
	ObjectID           string           `xml:"objectId"`
	ExtInfo            *FinderExtInfo   `xml:"extInfo"`
	MediaCount         int              `xml:"mediaCount"`
	MediaList          *FinderMediaList `xml:"mediaList"`
	ObjectNonceID      string           `xml:"objectNonceId"`
	BizUsernameV2      string           `xml:"bizUsernameV2"`
	AuthIconType       int              `xml:"authIconType"`
	Avatar             string           `xml:"avatar"`
	BizNickname        string           `xml:"bizNickname"`
	Username           string           `xml:"username"`
	BizAuthIconURL     string           `xml:"bizAuthIconUrl"`
	AuthIconURL        string           `xml:"authIconUrl"`
	SourceCommentScene int              `xml:"sourceCommentScene"`
	BizAvatar          string           `xml:"bizAvatar"`
}

type FinderExtInfo struct {
	TabContextID  string `xml:"tabContextId"`
	ContextID     string `xml:"contextId"`
	ShareSrcScene int    `xml:"shareSrcScene"`
}

type FinderMediaList struct {
	Media []FinderMedia `xml:"media"`
}

type FinderMedia struct {
	Height            float64 `xml:"height"`
	MediaType         int     `xml:"mediaType"`
	Width             float64 `xml:"width"`
	ThumbURL          string  `xml:"thumbUrl"`
	VideoPlayDuration int     `xml:"videoPlayDuration"`
	CoverURL          string  `xml:"coverUrl"`
	URL               string  `xml:"url"`
}
