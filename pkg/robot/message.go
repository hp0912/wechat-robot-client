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

type SystemMessage struct {
	XMLName   xml.Name  `xml:"sysmsg"`
	Type      string    `xml:"type,attr"`
	RevokeMsg RevokeMsg `xml:"revokemsg"`
}

type RevokeMsg struct {
	XMLName    xml.Name `xml:"revokemsg"`
	Session    string   `xml:"session"`
	MsgID      int64    `xml:"msgid"`
	NewMsgID   int64    `xml:"newmsgid"`
	ReplaceMsg string   `xml:"replacemsg"`
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

type MusicSearchResponse struct {
	Data MusicSearchData `json:"data"`
}

type MusicSearchData struct {
	Code   int     `json:"code"`
	Title  *string `json:"title"`
	Singer string  `json:"singer"`
	ID     string  `json:"id"`
	Cover  *string `json:"cover"`
	Link   string  `json:"link"`
	Url    string  `json:"url"`
	Lyric  *string `json:"lyric"`
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
