package robot

type LinkedinContactItem struct {
	LinkedinMemberId  string `json:"LinkedinMemberId"`
	LinkedinName      string `json:"LinkedinName"`
	LinkedinPublicUrl string `json:"LinkedinPublicUrl"`
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
	NickName BuiltinString `json:"NickName"`
	UserName BuiltinString `json:"UserName"`
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
	ChatRoomOwner         string                `json:"ChatRoomOwner"`
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
	ImgBuf                BuiltinBuffer         `json:"ImgBuf"`
	ImgFlag               int                   `json:"ImgFlag"`
	LabelIdList           string                `json:"LabelIdlist"`
	Level                 int                   `json:"Level"`
	MobileFullHash        string                `json:"MobileFullHash"`
	MobileHash            string                `json:"MobileHash"`
	MyBrandList           string                `json:"MyBrandList"`
	NewChatroomData       NewChatroomData       `json:"NewChatroomData"`
	NickName              BuiltinString         `json:"NickName"`
	PersonalCard          int                   `json:"PersonalCard"`
	PhoneNumListInfo      PhoneNumListInfo      `json:"PhoneNumListInfo"`
	Province              string                `json:"Province"`
	Pyinitial             BuiltinString         `json:"Pyinitial"`
	QuanPin               BuiltinString         `json:"QuanPin"`
	RealName              string                `json:"RealName"`
	Remark                BuiltinString         `json:"Remark"`
	RemarkPyinitial       BuiltinString         `json:"RemarkPyinitial"`
	RemarkQuanPin         BuiltinString         `json:"RemarkQuanPin"`
	RoomInfoCount         int                   `json:"RoomInfoCount"`
	RoomInfoList          []RoomInfo            `json:"RoomInfoList"`
	Sex                   int                   `json:"Sex"`
	Signature             string                `json:"Signature"`
	SmallHeadImgUrl       string                `json:"SmallHeadImgUrl"`
	SnsUserInfo           SnsUserInfo           `json:"SnsUserInfo"`
	Source                int                   `json:"Source"`
	UserName              BuiltinString         `json:"UserName"`
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
	DeleteContactScen int           `json:"DeleteContactScene"`
	UserName          BuiltinString `json:"UserName"`
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
		AntispamTicket string `json:"Antispamticket"`
		Username       string `json:"Username"`
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
