package robot

type LinkedinContactItem struct {
	LinkedinName      *string `protobuf:"bytes,1,opt,name=LinkedinName" json:"LinkedinName,omitempty"`
	LinkedinMemberId  *string `protobuf:"bytes,2,opt,name=LinkedinMemberId" json:"LinkedinMemberId,omitempty"`
	LinkedinPublicUrl *string `protobuf:"bytes,3,opt,name=LinkedinPublicUrl" json:"LinkedinPublicUrl,omitempty"`
}

type AdditionalContactList struct {
	LinkedinContactItem LinkedinContactItem `json:"LinkedinContactItem"`
}

type CustomizedInfo struct {
	BrandFlag    int    `json:"BrandFlag"`
	BrandIconURL string `json:"BrandIconURL"`
	BrandInfo    string `json:"BrandInfo"`
	ExternalInfo string `json:"ExternalInfo"`
}

type PhoneNumListInfo struct {
	Count        int      `json:"Count"`
	PhoneNumList []string `json:"PhoneNumList"`
}

type RoomInfo struct {
	NickName SKBuiltinStringT `json:"NickName"`
	UserName SKBuiltinStringT `json:"UserName"`
}

type Contact struct {
	AddContactScene       int                   `json:"AddContactScene"`
	AdditionalContactList AdditionalContactList `json:"AdditionalContactList"`
	AlbumBGImgID          string                `json:"AlbumBGImgID"`
	AlbumFlag             int                   `json:"AlbumFlag"`
	AlbumStyle            int                   `json:"AlbumStyle"`
	Alias                 string                `json:"Alias"`
	BigHeadImgUrl         string                `json:"BigHeadImgUrl"`
	BitMask               int                   `json:"BitMask"`
	BitVal                int                   `json:"BitVal"`
	CardImgUrl            string                `json:"CardImgUrl"`
	ChatRoomBusinessType  int                   `json:"chatRoomBusinessType"`
	ChatRoomData          string                `json:"ChatRoomData"`
	ChatRoomNotify        int                   `json:"ChatRoomNotify"`
	ChatRoomOwner         *string               `json:"ChatRoomOwner"`
	ChatroomAccessType    int                   `json:"ChatroomAccessType"`
	ChatroomInfoVersion   int                   `json:"ChatroomInfoVersion"`
	ChatroomMaxCount      int                   `json:"ChatroomMaxCount"`
	ChatroomStatus        int                   `json:"ChatroomStatus"`
	ChatroomVersion       int                   `json:"ChatroomVersion"`
	City                  string                `json:"City"`
	ContactType           int                   `json:"ContactType"`
	Country               string                `json:"Country"`
	CustomizedInfo        CustomizedInfo        `json:"CustomizedInfo"`
	DeleteFlag            int                   `json:"DeleteFlag"`
	DeleteContactScene    int                   `json:"DeleteContactScene"`
	Description           string                `json:"Description"`
	DomainList            any                   `json:"DomainList"`
	EncryptUserName       string                `json:"EncryptUserName"`
	ExtInfo               string                `json:"ExtInfo"`
	ExtFlag               int                   `json:"ExtFlag"`
	HasWeiXinHdHeadImg    int                   `json:"HasWeiXinHdHeadImg"`
	HeadImgMd5            string                `json:"HeadImgMd5"`
	IdCardNum             string                `json:"IdcardNum"`
	ImgBuf                SKBuiltinBufferT      `json:"ImgBuf"`
	ImgFlag               int                   `json:"ImgFlag"`
	LabelIdList           string                `json:"LabelIdlist"`
	Level                 int                   `json:"Level"`
	MobileFullHash        string                `json:"MobileFullHash"`
	MobileHash            string                `json:"MobileHash"`
	MyBrandList           string                `json:"MyBrandList"`
	NewChatroomData       NewChatroomData       `json:"NewChatroomData"`
	NickName              SKBuiltinStringT      `json:"NickName"`
	PersonalCard          int                   `json:"PersonalCard"`
	PhoneNumListInfo      PhoneNumListInfo      `json:"PhoneNumListInfo"`
	Province              string                `json:"Province"`
	Pyinitial             SKBuiltinStringT      `json:"Pyinitial"`
	QuanPin               SKBuiltinStringT      `json:"QuanPin"`
	RealName              string                `json:"RealName"`
	Remark                SKBuiltinStringT      `json:"Remark"`
	RemarkPyinitial       SKBuiltinStringT      `json:"RemarkPyinitial"`
	RemarkQuanPin         SKBuiltinStringT      `json:"RemarkQuanPin"`
	RoomInfoCount         int                   `json:"RoomInfoCount"`
	RoomInfoList          []RoomInfo            `json:"RoomInfoList"`
	Sex                   int                   `json:"Sex"`
	Signature             string                `json:"Signature"`
	SmallHeadImgUrl       string                `json:"SmallHeadImgUrl"`
	SnsUserInfo           SnsUserInfo           `json:"SnsUserInfo"`
	Source                int                   `json:"Source"`
	UserName              SKBuiltinStringT      `json:"UserName"`
	SourceExtInfo         string                `json:"SourceExtInfo"`
	VerifyContent         string                `json:"VerifyContent"`
	VerifyFlag            int                   `json:"VerifyFlag"`
	VerifyInfo            string                `json:"VerifyInfo"`
	WeiDianInfo           string                `json:"WeiDianInfo"`
	Weibo                 string                `json:"Weibo"`
	WeiboFlag             int                   `json:"WeiboFlag"`
	WeiboNickname         string                `json:"WeiboNickname"`
}

type DelContact struct {
	DeleteContactScen int              `json:"DeleteContactScene"`
	UserName          SKBuiltinStringT `json:"UserName"`
}

type MMBizJsApiGetUserOpenIdResponse struct {
	BaseResponse         *BaseResponse `json:"BaseResponse,omitempty"`
	Openid               *string       `json:"Openid,omitempty"`
	NickName             *string       `json:"NickName,omitempty"`
	HeadImgUrl           *string       `json:"HeadImgUrl,omitempty"`
	Sign                 *string       `json:"Sign,omitempty"`
	FriendRelation       *uint32       `json:"FriendRelation,omitempty"` // 1//删除 4/自己拉黑 5/被拉黑 0/正常
	XXX_NoUnkeyedLiteral *string       `json:"XXX_NoUnkeyedLiteral,omitempty"`
	XXXUnrecognized      *string       `json:"XXX_unrecognized,omitempty"`
	XXXSizecache         *string       `json:"XXX_sizecache,omitempty"`
}

type FriendSearchRequest struct {
	Wxid        string `json:"Wxid"`
	ToUserName  string `json:"ToUserName"`
	FromScene   int    `json:"FromScene"`
	SearchScene int    `json:"SearchScene"`
}

type SearchContactResponse struct {
	BaseResponse       *BaseResponse        `json:"BaseResponse,omitempty"`
	UserName           *SKBuiltinStringT    `json:"UserName,omitempty"`
	NickName           *SKBuiltinStringT    `json:"NickName,omitempty"`
	Pyinitial          *SKBuiltinStringT    `json:"Pyinitial,omitempty"`
	QuanPin            *SKBuiltinStringT    `json:"QuanPin,omitempty"`
	Sex                *int32               `json:"Sex,omitempty"`
	ImgBuf             *SKBuiltinBufferT    `json:"ImgBuf,omitempty"`
	Province           *string              `json:"Province,omitempty"`
	City               *string              `json:"City,omitempty"`
	Signature          *string              `json:"Signature,omitempty"`
	PersonalCard       *uint32              `json:"PersonalCard,omitempty"`
	VerifyFlag         *int32               `json:"VerifyFlag,omitempty"`
	VerifyInfo         *string              `json:"VerifyInfo,omitempty"`
	Weibo              *string              `json:"Weibo,omitempty"`
	Alias              *string              `json:"Alias,omitempty"`
	WeiboNickname      *string              `json:"WeiboNickname,omitempty"`
	WeiboFlag          *int32               `json:"WeiboFlag,omitempty"`
	AlbumStyle         *int32               `json:"AlbumStyle,omitempty"`
	AlbumFlag          *int32               `json:"AlbumFlag,omitempty"`
	AlbumBgimgId       *string              `json:"AlbumBgimgId,omitempty"`
	SnsUserInfo        *SnsUserInfo         `json:"SnsUserInfo,omitempty"`
	Country            *string              `json:"Country,omitempty"`
	MyBrandList        *string              `json:"MyBrandList,omitempty"`
	CustomizedInfo     *CustomizedInfo      `json:"CustomizedInfo,omitempty"`
	ContactCount       *uint32              `json:"ContactCount,omitempty"`
	Contactlist        []*SearchContactItem `json:"Contactlist,omitempty"`
	BigHeadImgUrl      *string              `json:"BigHeadImgUrl,omitempty"`
	SmallHeadImgUrl    *string              `json:"SmallHeadImgUrl,omitempty"`
	ResBuf             *SKBuiltinBufferT    `json:"ResBuf,omitempty"`
	AntispamTicket     *string              `json:"AntispamTicket,omitempty"`
	KfworkerId         *string              `json:"KfworkerId,omitempty"`
	MatchType          *uint32              `json:"MatchType,omitempty"`
	PopupInfoMsg       *string              `json:"PopupInfoMsg,omitempty"`
	OpenImcontactCount *uint32              `json:"OpenImcontactCount,omitempty"`
	OpenImcontactList  []*OpenIMContact     `json:"OpenImcontactList,omitempty"`
}

type SearchContactItem struct {
	UserName        *SKBuiltinStringT `json:"UserName,omitempty"`
	NickName        *SKBuiltinStringT `json:"NickName,omitempty"`
	Pyinitial       *SKBuiltinStringT `json:"Pyinitial,omitempty"`
	QuanPin         *SKBuiltinStringT `json:"QuanPin,omitempty"`
	Sex             *int32            `json:"sex,omitempty"`
	ImgBuf          *SKBuiltinBufferT `json:"imgBuf,omitempty"`
	Province        *string           `json:"province,omitempty"`
	City            *string           `json:"city,omitempty"`
	Signature       *string           `json:"signature,omitempty"`
	PersonalCard    *uint32           `json:"personalCard,omitempty"`
	VerifyFlag      *uint32           `json:"verifyFlag,omitempty"`
	VerifyInfo      *string           `json:"verifyInfo,omitempty"`
	Weibo           *string           `json:"weibo,omitempty"`
	Alias           *string           `json:"alias,omitempty"`
	WeiboNickname   *string           `json:"weiboNickname,omitempty"`
	WeiboFlag       *uint32           `json:"weiboFlag,omitempty"`
	AlbumStyle      *int32            `json:"albumStyle,omitempty"`
	AlbumFlag       *int32            `json:"albumFlag,omitempty"`
	AlbumBgimgId    *string           `json:"albumBgimgId,omitempty"`
	SnsUserInfo     *SnsUserInfo      `json:"snsUserInfo,omitempty"`
	Country         *string           `json:"country,omitempty"`
	MyBrandList     *string           `json:"myBrandList,omitempty"`
	CustomizedInfo  *CustomizedInfo   `json:"customizedInfo,omitempty"`
	BigHeadImgUrl   *string           `json:"bigHeadImgUrl,omitempty"`
	SmallHeadImgUrl *string           `json:"smallHeadImgUrl,omitempty"`
	AntispamTicket  *string           `json:"antispamTicket,omitempty"`
	MatchType       *uint32           `json:"matchType,omitempty"`
}

type OpenIMContact struct {
	TpUserName      *string                  `json:"tpUserName,omitempty"`
	Nickname        *string                  `json:"nickname,omitempty"`
	Type            *uint32                  `json:"type,omitempty"`
	Remark          *string                  `json:"remark,omitempty"`
	BigHeadimg      *string                  `json:"bigHeadimg,omitempty"`
	SmallHeadimg    *string                  `json:"smallHeadimg,omitempty"`
	Source          *uint32                  `json:"source,omitempty"`
	NicknamePyinit  *string                  `json:"nicknamePyinit,omitempty"`
	NicknameQuanpin *string                  `json:"nicknameQuanpin,omitempty"`
	RemarkPyinit    *string                  `json:"remarkPyinit,omitempty"`
	RemarkQuanpin   *string                  `json:"remarkQuanpin,omitempty"`
	CustomInfo      *OpenIMContactCustomInfo `json:"customInfo,omitempty"`
	AntispamTicket  *string                  `json:"antispamTicket,omitempty"`
	AppId           *string                  `json:"appId,omitempty"`
	Sex             *uint32                  `json:"sex,omitempty"`
	DescWordingId   *string                  `json:"descWordingId,omitempty"`
}

type OpenIMContactCustomInfo struct {
	DetailVisible *uint32 `json:"detailVisible,omitempty"`
	Tdetail       *string `json:"tdetail,omitempty"`
}

type FriendSendRequestParam struct {
	Wxid          string `json:"Wxid"`
	Opcode        int    `json:"Opcode"` // 1免验证发送请求, 2发送验证申请, 3通过好友验证
	Scene         int    `json:"Scene"`  // 1来源QQ，2来源邮箱，3来源微信号，14群聊，15手机号，18附近的人，25漂流瓶，29摇一摇，30二维码，13来源通讯录
	V1            string `json:"V1"`
	V2            string `json:"V2"`
	VerifyContent string `json:"VerifyContent"`
}

type FriendSetRemarksRequest struct {
	Wxid    string `json:"Wxid"`
	ToWxid  string `json:"ToWxid"`
	Remarks string `json:"Remarks"`
}

type GetContactListResponse struct {
	BaseResponse              BaseResponse `json:"BaseResponse"`
	CurrentWxcontactSeq       int          `json:"CurrentWxcontactSeq"`
	CurrentChatRoomContactSeq int          `json:"CurrentChatRoomContactSeq"`
	CountinueFlag             int          `json:"CountinueFlag"`
	ContactUsernameList       []string     `json:"ContactUsernameList"` // 联系人微信Id列表
}

type GetContactResponse struct {
	BaseResponse BaseResponse `json:"BaseResponse"`
	ContactCount int          `json:"ContactCount"`
	ContactList  []Contact    `json:"ContactList"`
	Ret          []int        `json:"Ret"`
	Ticket       []struct {
		AntispamTicket *string `json:"Antispamticket"`
		Username       *string `json:"Username"`
	} `json:"Ticket"`
	SendMsgTicketList [][]int `json:"sendMsgTicketList"`
}

type GetContactListRequest struct {
	Wxid                      string `json:"Wxid"`
	CurrentChatRoomContactSeq int    `json:"CurrentChatRoomContactSeq"`
	CurrentWxcontactSeq       int    `json:"CurrentWxcontactSeq"`
}

type GetContactDetailRequest struct {
	Wxid     string `json:"Wxid"`
	Towxids  string `json:"Towxids"`
	ChatRoom string `json:"ChatRoom"`
}

type FriendPassVerifyRequest struct {
	Wxid  string `json:"Wxid"`
	Scene int    `json:"Scene"`
	V1    string `json:"V1"`
	V2    string `json:"V2"`
}

type VerifyUserResponse struct {
	BaseResponse *BaseResponse `json:"BaseResponse,omitempty"`
	Username     *string       `json:"Username,omitempty"`
}

type FriendDeleteRequest struct {
	Wxid   string `json:"Wxid"`
	ToWxid string `json:"ToWxid"`
}
