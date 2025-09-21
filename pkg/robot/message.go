package robot

import (
	"encoding/xml"
	"wechat-robot-client/model"
)

type SyncMessage struct {
	ModUserInfos    []*UserInfo       `json:"ModUserInfos"`
	ModContacts     []*Contact        `json:"ModContacts"`
	DelContacts     []*DelContact     `json:"DelContacts"`
	ModUserImgs     []*UserImg        `json:"ModUserImgs"`
	FunctionSwitchs []*FunctionSwitch `json:"FunctionSwitchs"`
	UserInfoExts    []*UserInfoExt    `json:"UserInfoExts"`
	AddMsgs         []Message         `json:"AddMsgs"`
	AddSnsBuffer    []string          `json:"AddSnsBuffer"`
	ContinueFlag    int               `json:"ContinueFlag"`
	KeyBuf          SKBuiltinBufferT  `json:"KeyBuf"`
	Status          int               `json:"Status"`
	Continue        int               `json:"Continue"`
	Time            int               `json:"Time"`
	UnknownCmdId    string            `json:"UnknownCmdId"`
	Remarks         string            `json:"Remarks"`
}

type Message struct {
	MsgId        int64             `json:"MsgId"`
	FromUserName SKBuiltinStringT  `json:"FromUserName"`
	ToUserName   SKBuiltinStringT  `json:"ToUserName"`
	Content      SKBuiltinStringT  `json:"Content"`
	CreateTime   int64             `json:"CreateTime"`
	MsgType      model.MessageType `json:"MsgType"`
	Status       int               `json:"Status"`
	ImgStatus    int               `json:"ImgStatus"`
	ImgBuf       SKBuiltinBufferT  `json:"ImgBuf"`
	MsgSource    string            `json:"MsgSource"`
	NewMsgId     int64             `json:"NewMsgId"`
	MsgSeq       int               `json:"MsgSeq"`
	PushContent  string            `json:"PushContent,omitempty"`
}

type FunctionSwitch struct {
	FunctionId  int64 `json:"FunctionId"`
	SwitchValue int64 `json:"SwitchValue"`
}

type SyncMessageRequest struct {
	Wxid    string `json:"Wxid"`
	Scene   int    `json:"Scene"`
	Synckey string `json:"Synckey"`
}

type NewFriendMessage struct {
	XMLName           xml.Name  `xml:"msg"`
	FromUsername      string    `xml:"fromusername,attr"`
	EncryptUsername   string    `xml:"encryptusername,attr"`
	FromNickname      string    `xml:"fromnickname,attr"`
	Content           string    `xml:"content,attr"`
	FullPy            string    `xml:"fullpy,attr"`
	ShortPy           string    `xml:"shortpy,attr"`
	ImageStatus       string    `xml:"imagestatus,attr"`
	Scene             string    `xml:"scene,attr"`
	Country           string    `xml:"country,attr"`
	Province          string    `xml:"province,attr"`
	City              string    `xml:"city,attr"`
	Sign              string    `xml:"sign,attr"`
	PerCard           string    `xml:"percard,attr"`
	Sex               string    `xml:"sex,attr"`
	Alias             string    `xml:"alias,attr"`
	Weibo             string    `xml:"weibo,attr"`
	AlbumFlag         string    `xml:"albumflag,attr"`
	AlbumStyle        string    `xml:"albumstyle,attr"`
	AlbumBgImgID      string    `xml:"albumbgimgid,attr"`
	SnsFlag           string    `xml:"snsflag,attr"`
	SnsBgImgID        string    `xml:"snsbgimgid,attr"`
	SnsBgObjectID     string    `xml:"snsbgobjectid,attr"`
	MHash             string    `xml:"mhash,attr"`
	MFullHash         string    `xml:"mfullhash,attr"`
	BigHeadImgURL     string    `xml:"bigheadimgurl,attr"`
	SmallHeadImgURL   string    `xml:"smallheadimgurl,attr"`
	Ticket            string    `xml:"ticket,attr"`
	OpCode            string    `xml:"opcode,attr"`
	GoogleContact     string    `xml:"googlecontact,attr"`
	QrTicket          string    `xml:"qrticket,attr"`
	ChatroomUsername  string    `xml:"chatroomusername,attr"`
	SourceUsername    string    `xml:"sourceusername,attr"`
	SourceNickname    string    `xml:"sourcenickname,attr"`
	ShareCardUsername string    `xml:"sharecardusername,attr"`
	ShareCardNickname string    `xml:"sharecardnickname,attr"`
	CardVersion       string    `xml:"cardversion,attr"`
	ExtFlag           string    `xml:"extflag,attr"`
	BrandList         BrandList `xml:"brandlist"`
}

type BrandList struct {
	XMLName xml.Name `xml:"brandlist"`
	Count   string   `xml:"count,attr"`
	Ver     string   `xml:"ver,attr"`
}

type XmlMessage struct {
	XMLName      xml.Name   `xml:"msg"`
	AppMsg       AppMessage `xml:"appmsg"`
	FromUsername string     `xml:"fromusername"`
	Scene        int        `xml:"scene"`
	AppInfo      AppInfo    `xml:"appinfo"`
	CommentURL   string     `xml:"commenturl"`
}

type AppMessage struct {
	AppID             string       `xml:"appid,attr"`
	SDKVer            string       `xml:"sdkver,attr"`
	Title             string       `xml:"title"`
	Des               string       `xml:"des"`
	Action            string       `xml:"action"`
	Type              int          `xml:"type"`
	ShowType          int          `xml:"showtype"`
	SoundType         int          `xml:"soundtype"`
	MediaTagName      string       `xml:"mediatagname"`
	MessageExt        string       `xml:"messageext"`
	MessageAction     string       `xml:"messageaction"`
	Content           string       `xml:"content"`
	ContentAttr       int          `xml:"contentattr"`
	URL               string       `xml:"url"`
	LowURL            string       `xml:"lowurl"`
	DataURL           string       `xml:"dataurl"`
	LowDataURL        string       `xml:"lowdataurl"`
	SongAlbumURL      string       `xml:"songalbumurl"`
	SongLyric         string       `xml:"songlyric"`
	AppAttach         AppAttach    `xml:"appattach"`
	ExtInfo           string       `xml:"extinfo"`
	SourceUsername    string       `xml:"sourceusername"`
	SourceDisplayName string       `xml:"sourcedisplayname"`
	ThumbURL          string       `xml:"thumburl"`
	MD5               string       `xml:"md5"`
	StatExtStr        string       `xml:"statextstr"`
	ReferMsg          ReferMessage `xml:"refermsg"`
}

type ReferMessage struct {
	Type        int    `xml:"type"`
	SvrID       string `xml:"svrid"`
	FromUsr     string `xml:"fromusr"`
	ChatUsr     string `xml:"chatusr"`
	DisplayName string `xml:"displayname"`
	Content     string `xml:"content"`
	MsgSource   string `xml:"msgsource"`
	CreateTime  int64  `xml:"createtime"`
}

type SystemMessage struct {
	XMLName        xml.Name       `xml:"sysmsg"`
	Type           string         `xml:"type,attr"`
	RevokeMsg      RevokeMsg      `xml:"revokemsg"`
	Pat            Pat            `xml:"pat,omitempty"`
	SysMsgTemplate SysMsgTemplate `xml:"sysmsgtemplate"`
}

type RevokeMsg struct {
	XMLName    xml.Name `xml:"revokemsg"`
	Session    string   `xml:"session"`
	MsgID      int64    `xml:"msgid"`
	NewMsgID   int64    `xml:"newmsgid"`
	ReplaceMsg string   `xml:"replacemsg"`
}

type SysMsgTemplate struct {
	ContentTemplate ContentTemplate `xml:"content_template"`
}

type ContentTemplate struct {
	Type     string   `xml:"type,attr"`
	Plain    string   `xml:"plain"`
	Template string   `xml:"template"`
	LinkList LinkList `xml:"link_list"`
}

type LinkList struct {
	Links []Link `xml:"link"`
}

type Link struct {
	Name         string        `xml:"name,attr"`
	Type         string        `xml:"type,attr"`
	Hidden       string        `xml:"hidden,attr,omitempty"`
	MemberList   *MemberList   `xml:"memberlist,omitempty"`
	Separator    string        `xml:"separator,omitempty"`
	Title        string        `xml:"title,omitempty"`
	UsernameList *UsernameList `xml:"usernamelist,omitempty"`
}

type Pat struct {
	XMLName          xml.Name `xml:"pat"`
	FromUsername     string   `xml:"fromusername"`
	ChatUsername     string   `xml:"chatusername"`
	PattedUsername   string   `xml:"pattedusername"`
	PatSuffix        string   `xml:"patsuffix"`
	PatSuffixVersion int      `xml:"patsuffixversion"`
	Template         string   `xml:"template"`
}

type MemberList struct {
	Members []Member `xml:"member"`
}

type Member struct {
	Username string `xml:"username"`
	Nickname string `xml:"nickname"`
}

type UsernameList struct {
	Usernames []string `xml:"username"`
}

type MessageRevokeRequest struct {
	Wxid        string `json:"Wxid"`
	ClientMsgId int64  `json:"ClientMsgId"`
	NewMsgId    int64  `json:"NewMsgId"`
	ToUserName  string `json:"ToUserName"`
	CreateTime  int64  `json:"CreateTime"`
}

type MessageRevokeResponse struct {
	BaseResponse
	IsysWording string `json:"isysWording"`
}

type SendTextMessageRequest struct {
	Wxid    string `json:"Wxid"`
	Type    int    `json:"Type"`
	ToWxid  string `json:"ToWxid"`
	Content string `json:"Content"`
	At      string `json:"At"`
}

type MsgSendGroupMassMsgTextRequest struct {
	Wxid    string
	ToWxid  []string
	Content string
}

type TextMessageResponse struct {
	Ret         int              `json:"Ret"`
	ToUsetName  SKBuiltinStringT `json:"ToUsetName"`
	MsgId       int64            `json:"MsgId"`
	ClientMsgid int64            `json:"ClientMsgid"`
	Createtime  int64            `json:"Createtime"`
	Servertime  int64            `json:"servertime"`
	Type        int              `json:"Type"`
	NewMsgId    int64            `json:"NewMsgId"`
}

type SendTextMessageResponse struct {
	BaseResponse
	List   []TextMessageResponse `json:"List"`
	Count  int                   `json:"Count"`
	NoKnow int                   `json:"NoKnow"`
}

type MsgSendGroupMassMsgTextResponse struct {
	BaseResponse  *BaseResponse `json:"baseResponse,omitempty"`
	DataStartPos  *uint32       `json:"dataStartPos,omitempty"`
	ThumbStartPos *uint32       `json:"thumbStartPos,omitempty"`
	MaxSupport    *uint32       `json:"maxSupport,omitempty"`
}

type MsgUploadImgRequest struct {
	Wxid   string `json:"Wxid"`
	ToWxid string `json:"ToWxid"`
	Base64 string `json:"Base64"`
}

type MsgUploadImgResponse struct {
	BaseResponse
	Msgid        int64            `json:"Msgid"`
	ClientImgId  SKBuiltinStringT `json:"ClientImgId"`
	FromUserName SKBuiltinStringT `json:"FromUserName"`
	ToUserName   SKBuiltinStringT `json:"ToUserName"`
	TotalLen     int64            `json:"TotalLen"`
	StartPos     int64            `json:"StartPos"`
	DataLen      int64            `json:"DataLen"`
	CreateTime   int64            `json:"CreateTime"`
	Newmsgid     int64            `json:"Newmsgid"`
	MsgSource    string           `json:"MsgSource"`
}

type MsgSendVideoRequest struct {
	Wxid        string `json:"Wxid"`
	ToWxid      string `json:"ToWxid"`
	Base64      string `json:"Base64"`
	ImageBase64 string `json:"ImageBase64"`
	PlayLength  int64  `json:"PlayLength"`
}

type MsgSendVideoResponse struct {
	BaseResponse
	Msgid         int64  `json:"msgId"`
	ClientMsgId   string `json:"clientMsgId"`
	ThumbStartPos int64  `json:"thumbStartPos"`
	VideoStartPos int64  `json:"videoStartPos"`
	NewMsgId      int64  `json:"newMsgId"`
	ActionFlag    int    `json:"actionFlag"`
}

type MsgSendVoiceRequest struct {
	Wxid      string `json:"Wxid"`
	ToWxid    string `json:"ToWxid"`
	Type      int    `json:"Type"`
	Base64    string `json:"Base64"`
	VoiceTime int    `json:"VoiceTime"`
}

type MsgSendVoiceResponse struct {
	BaseResponse
	NewMsgId     int64  `json:"NewMsgId"`
	MsgId        int64  `json:"MsgId"`
	ClientMsgId  string `json:"ClientMsgId"`
	FromUserName string `json:"FromUserName"`
	ToUserName   string `json:"ToUserName"`
	Offset       int    `json:"Offset"`
	Length       int    `json:"Length"`
	VoiceLength  int    `json:"VoiceLength"`
	EndFlag      int    `json:"EndFlag"`
	CancelFlag   int    `json:"CancelFlag"`
	CreateTime   int64  `json:"CreateTime"`
}

type AppMessageCommon struct {
	FromUsername string `json:"FromUsername"`
}

type SendAppRequest struct {
	Wxid   string `json:"Wxid"`
	ToWxid string `json:"ToWxid"`
	Type   int    `json:"Type"`
	Xml    string `json:"Xml"`
}

type SendAppResponse struct {
	FromUserName string `json:"fromUserName"`
	Type         int    `json:"type"`
	ActionFlag   int    `json:"actionFlag"`
	ToUserName   string `json:"toUserName"`
	MsgId        int64  `json:"msgId"`
	ClientMsgId  string `json:"clientMsgId"`
	CreateTime   int64  `json:"createTime"`
	NewMsgId     int64  `json:"newMsgId"`
	MsgSource    string `json:"msgSource"`
	Content      string `json:"content"`
}

type SongInfo struct {
	AppMessageCommon
	AppID    string `json:"AppID"`
	Title    string `json:"Title"`
	Singer   string `json:"Singer"`
	Url      string `json:"Url"`
	MusicUrl string `json:"MusicUrl"`
	CoverUrl string `json:"CoverUrl"`
	Lyric    string `json:"Lyric"`
}

type ShareLinkInfo struct {
	Title    string `json:"Title"`
	Desc     string `json:"Desc"`
	Url      string `json:"Url"`
	ThumbUrl string `json:"ThumbUrl"`
}

type MusicSearchResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data MusicSearchData `json:"data"`
}

type MusicSearchData struct {
	Title    *string `json:"title"`
	Singer   string  `json:"singer"`
	ID       string  `json:"id"`
	Cover    *string `json:"cover"`
	Link     string  `json:"link"`
	MusicURL string  `json:"music_url"`
	Lrc      *string `json:"lrc"`
}

type SendEmojiRequest struct {
	Wxid     string `json:"Wxid"`
	ToWxid   string `json:"ToWxid"`
	Md5      string `json:"Md5"`
	TotalLen int32  `json:"TotalLen"`
}

type EmojiItem struct {
	Ret      int    `json:"ret"`
	StartPos int    `json:"startPos"`
	TotalLen int    `json:"totalLen"`
	Md5      string `json:"md5"`
	MsgId    int64  `json:"msgId"`
	NewMsgId int64  `json:"newMsgId"`
}

type SendEmojiResponse struct {
	BaseResponse
	EmojiItemCount int         `json:"emojiItemCount"`
	ActionFlag     int64       `json:"actionFlag"`
	EmojiItem      []EmojiItem `json:"emojiItem"`
}

type ShareLinkRequest struct {
	Wxid   string `json:"Wxid"`
	ToWxid string `json:"ToWxid"`
	Type   int32  `json:"Type"`
	Xml    string `json:"Xml"`
}

type ShareLinkResponse struct {
	BaseResponse
	FromUserName string `json:"fromUserName"`
	Type         int    `json:"type"`
	ActionFlag   int    `json:"actionFlag"`
	ToUserName   string `json:"toUserName"`
	MsgId        int64  `json:"msgId"`
	ClientMsgId  string `json:"clientMsgId"`
	CreateTime   int64  `json:"createTime"`
	NewMsgId     int64  `json:"newMsgId"`
	MsgSource    string `json:"msgSource"`
}

type SendCDNAttachmentRequest struct {
	Wxid    string `json:"Wxid"`
	ToWxid  string `json:"ToWxid"`
	Content string `json:"Content"`
}

type SendCDNFileResponse struct {
	BaseResponse
	ToUserName   string `json:"toUserName"`
	ClientMsgId  string `json:"clientMsgId"`
	Type         int    `json:"type"`
	NewMsgId     int64  `json:"newMsgId"`
	MsgSource    string `json:"msgSource"`
	ActionFlag   int    `json:"actionFlag"`
	FromUserName string `json:"fromUserName"`
	MsgId        int64  `json:"msgId"`
	CreateTime   int64  `json:"createTime"`
	Aeskey       string `json:"aeskey"`
}

type SendCDNImgResponse struct {
	BaseResponse
	FromUserName SKBuiltinStringT `json:"FromUserName"`
	DataLen      int64            `json:"DataLen"`
	CreateTime   int64            `json:"CreateTime"`
	Newmsgid     int64            `json:"Newmsgid"`
	Fileid       string           `json:"Fileid"`
	MsgSource    string           `json:"MsgSource"`
	Msgid        int64            `json:"Msgid"`
	ClientImgId  SKBuiltinStringT `json:"ClientImgId"`
	ToUserName   SKBuiltinStringT `json:"ToUserName"`
	TotalLen     int64            `json:"TotalLen"`
	StartPos     int64            `json:"StartPos"`
	Aeskey       string           `json:"Aeskey"`
}

type SendCDNVideoResponse struct {
	BaseResponse
	ClientMsgId   string `json:"clientMsgId"`
	MsgId         int64  `json:"msgId"`
	VideoStartPos int64  `json:"videoStartPos"`
	NewMsgId      int64  `json:"newMsgId"`
	Aeskey        string `json:"aeskey"`
	MsgSource     string `json:"msgSource"`
	ActionFlag    int    `json:"actionFlag"`
	ThumbStartPos int64  `json:"thumbStartPos"`
}

type SendFileMessageRequest struct {
	Wxid            string `json:"Wxid"`
	ToWxid          string `json:"ToWxid"`
	ClientAppDataId string `json:"ClientAppDataId"`
	Filename        string `json:"Filename"`
	FileMD5         string `json:"FileMD5"`
	TotalLen        int64  `json:"TotalLen"`
	StartPos        int64  `json:"StartPos"`
	TotalChunks     int64  `json:"TotalChunks"`
}

type SendFileMessageResponse struct {
	BaseResponse    *BaseResponse `json:"BaseResponse,omitempty"`
	AppId           *string       `json:"appId,omitempty"`
	MediaId         *string       `json:"mediaId,omitempty"`
	ClientAppDataId *string       `json:"clientAppDataId,omitempty"`
	UserName        *string       `json:"userName,omitempty"`
	TotalLen        *uint32       `json:"totalLen,omitempty"`
	StartPos        *uint32       `json:"startPos,omitempty"`
	DataLen         *uint32       `json:"dataLen,omitempty"`
	CreateTime      *uint64       `json:"createTime,omitempty"`
}
