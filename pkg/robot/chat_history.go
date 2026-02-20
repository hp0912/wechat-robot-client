package robot

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type ChatHistoryMessage struct {
	XMLName      xml.Name           `xml:"msg"`
	AppMsg       ChatHistoryAppMsg  `xml:"appmsg"`
	FromUsername string             `xml:"fromusername,omitempty"`
	Scene        int                `xml:"scene,omitempty"`
	AppInfo      ChatHistoryAppInfo `xml:"appinfo,omitempty"`
	CommentURL   string             `xml:"commenturl,omitempty"`
}

type ChatHistoryMessageRecord struct {
	Nickname string `json:"nickname"`
	Content  string `json:"content"`
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
	// XML holds the raw inner XML of <recorditem>.
	// It is primarily used when marshaling (so callers can inject CDATA as-is).
	XML string `xml:",innerxml"`
	// Text holds the decoded text content of <recorditem> (CDATA becomes plain text).
	Text string `xml:"-"`
}

func (r *ChatHistoryRecordItem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var aux struct {
		Inner string `xml:",innerxml"`
		Text  string `xml:",chardata"`
	}
	if err := d.DecodeElement(&aux, &start); err != nil {
		return err
	}

	r.XML = aux.Inner
	r.Text = strings.TrimSpace(aux.Text)
	if r.Text == "" {
		r.Text = strings.TrimSpace(extractFirstCDATA(aux.Inner))
		if r.Text == "" {
			r.Text = strings.TrimSpace(aux.Inner)
		}
	}
	return nil
}

// ParseRecordInfo parses the payload inside <recorditem> into a RecordInfo tree.
// The payload is usually XML inside CDATA, e.g. "<recordinfo>...</recordinfo>".
// If the payload is empty or "(null)", it returns (nil, nil).
func (r ChatHistoryRecordItem) ParseRecordInfo() (*RecordInfo, error) {
	payload := strings.TrimSpace(r.Text)
	if payload == "" {
		payload = strings.TrimSpace(extractFirstCDATA(r.XML))
		if payload == "" {
			payload = strings.TrimSpace(r.XML)
		}
	}

	if payload == "" || payload == "(null)" {
		return nil, nil
	}
	if !strings.Contains(payload, "<") {
		return nil, fmt.Errorf("recorditem payload does not look like XML: %q", payload)
	}

	var recordInfo RecordInfo
	if err := xml.Unmarshal([]byte(payload), &recordInfo); err != nil {
		return nil, fmt.Errorf("unmarshal recorditem payload: %w", err)
	}
	return &recordInfo, nil
}

func extractFirstCDATA(s string) string {
	const start = "<![CDATA["
	const end = "]]>"
	startIdx := strings.Index(s, start)
	if startIdx < 0 {
		return ""
	}
	startIdx += len(start)
	endIdx := strings.Index(s[startIdx:], end)
	if endIdx < 0 {
		return ""
	}
	return s[startIdx : startIdx+endIdx]
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

func ExtractChatHistoryMessageRecords(recordInfo *RecordInfo) []ChatHistoryMessageRecord {
	if recordInfo == nil || len(recordInfo.DataList.Items) == 0 {
		return nil
	}

	records := make([]ChatHistoryMessageRecord, 0)
	var walk func(items []DataItem)
	walk = func(items []DataItem) {
		for _, item := range items {
			if item.DataType == 1 || item.DataType == 17 {
				nickname := strings.TrimSpace(item.SourceName)
				content := strings.TrimSpace(item.DataDesc)
				if nickname != "" || content != "" {
					records = append(records, ChatHistoryMessageRecord{
						Nickname: nickname,
						Content:  content,
					})
				}
			}

			if item.DataType == 17 && item.RecordXML != nil {
				nestedItems := item.RecordXML.RecordInfo.DataList.Items
				if len(nestedItems) > 0 {
					walk(nestedItems)
				}
			}
		}
	}

	walk(recordInfo.DataList.Items)
	return records
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
