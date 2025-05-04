package robot

import "wechat-robot-client/model"

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
	CreateTime   int               `json:"CreateTime"`
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

type MessageCommonXml struct {
	AesKey       string `xml:"aeskey,attr"`
	CdnMidImgUrl string `xml:"cdnmidimgurl,attr"`
	Length       int64  `xml:"length,attr"`
	Md5          string `xml:"md5,attr"`
}

type ImageMessageXml struct {
	Img MessageCommonXml `xml:"img"`
}
