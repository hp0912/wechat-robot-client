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
	KeyBuf          BuiltinBuffer     `json:"KeyBuf"`
	Status          int               `json:"Status"`
	Continue        int               `json:"Continue"`
	Time            int               `json:"Time"`
	UnknownCmdId    string            `json:"UnknownCmdId"`
	Remarks         string            `json:"Remarks"`
}

type Message struct {
	MsgId        int64             `json:"MsgId"`
	FromUserName BuiltinString     `json:"FromUserName"`
	ToUserName   BuiltinString     `json:"ToUserName"`
	Content      BuiltinString     `json:"Content"`
	CreateTime   int64             `json:"CreateTime"`
	MsgType      model.MessageType `json:"MsgType"`
	Status       int               `json:"Status"`
	ImgStatus    int               `json:"ImgStatus"`
	ImgBuf       BuiltinBuffer     `json:"ImgBuf"`
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
	Ret         int           `json:"Ret"`
	ToUsetName  BuiltinString `json:"ToUsetName"`
	MsgId       int64         `json:"MsgId"`
	ClientMsgid int64         `json:"ClientMsgid"`
	Createtime  int64         `json:"Createtime"`
	Servertime  int64         `json:"servertime"`
	Type        int           `json:"Type"`
	NewMsgId    int64         `json:"NewMsgId"`
}

type SendTextMessageResponse struct {
	BaseResponse
	List   []TextMessageResponse `json:"List"`
	Count  int                   `json:"Count"`
	NoKnow int                   `json:"NoKnow"`
}
