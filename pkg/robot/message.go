package robot

// MessageType 以Go惯用形式定义了PC微信所有的官方消息类型。
type MessageType int

// AppMessageType 以Go惯用形式定义了PC微信所有的官方App消息类型。
type AppMessageType int

const (
	MsgTypeText           MessageType = 1     // 文本消息
	MsgTypeImage          MessageType = 3     // 图片消息
	MsgTypeVoice          MessageType = 34    // 语音消息
	MsgTypeVerify         MessageType = 37    // 认证消息
	MsgTypePossibleFriend MessageType = 40    // 好友推荐消息
	MsgTypeShareCard      MessageType = 42    // 名片消息
	MsgTypeVideo          MessageType = 43    // 视频消息
	MsgTypeEmoticon       MessageType = 47    // 表情消息
	MsgTypeLocation       MessageType = 48    // 地理位置消息
	MsgTypeApp            MessageType = 49    // APP消息
	MsgTypeVoip           MessageType = 50    // VOIP消息
	MsgTypeInit           MessageType = 51    // 微信初始化消息
	MsgTypeVoipNotify     MessageType = 52    // VOIP结束消息
	MsgTypeVoipInvite     MessageType = 53    // VOIP邀请
	MsgTypeMicroVideo     MessageType = 62    // 小视频消息
	MsgTypeSys            MessageType = 10000 // 系统消息
	MsgTypeRecalled       MessageType = 10002 // 消息撤回
)

const (
	AppMsgTypeText                  AppMessageType = 1      // 文本消息
	AppMsgTypeImg                   AppMessageType = 2      // 图片消息
	AppMsgTypeAudio                 AppMessageType = 3      // 语音消息
	AppMsgTypeVideo                 AppMessageType = 4      // 视频消息
	AppMsgTypeUrl                   AppMessageType = 5      // 文章消息
	AppMsgTypeAttach                AppMessageType = 6      // 附件消息
	AppMsgTypeOpen                  AppMessageType = 7      // Open
	AppMsgTypeEmoji                 AppMessageType = 8      // 表情消息
	AppMsgTypeVoiceRemind           AppMessageType = 9      // VoiceRemind
	AppMsgTypeScanGood              AppMessageType = 10     // ScanGood
	AppMsgTypeGood                  AppMessageType = 13     // Good
	AppMsgTypeEmotion               AppMessageType = 15     // Emotion
	AppMsgTypeCardTicket            AppMessageType = 16     // 名片消息
	AppMsgTypeRealtimeShareLocation AppMessageType = 17     // 地理位置消息
	AppMsgTypeTransfers             AppMessageType = 2000   // 转账消息
	AppMsgTypeRedEnvelopes          AppMessageType = 2001   // 红包消息
	AppMsgTypeReaderType            AppMessageType = 100001 //自定义的消息
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
	MsgId        int64         `json:"MsgId"`
	FromUserName BuiltinString `json:"FromUserName"`
	ToUserName   BuiltinString `json:"ToUserName"`
	Content      BuiltinString `json:"Content"`
	CreateTime   int           `json:"CreateTime"`
	MsgType      MessageType   `json:"MsgType"`
	Status       int           `json:"Status"`
	ImgStatus    int           `json:"ImgStatus"`
	ImgBuf       BuiltinBuffer `json:"ImgBuf"`
	MsgSource    string        `json:"MsgSource"`
	NewMsgId     int64         `json:"NewMsgId"`
	MsgSeq       int           `json:"MsgSeq"`
	PushContent  string        `json:"PushContent,omitempty"`
}

type FunctionSwitch struct {
	FunctionId  int64 `json:"FunctionId"`
	SwitchValue int64 `json:"SwitchValue"`
}
