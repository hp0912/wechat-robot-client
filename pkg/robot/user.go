package robot

type UserInfo struct {
	AlbumBgimgId   string         `json:"AlbumBgimgId"`
	AlbumFlag      int            `json:"AlbumFlag"`
	AlbumStyle     int            `json:"AlbumStyle"`
	Alias          string         `json:"Alias"`
	BindEmail      BuiltinString  `json:"BindEmail"`
	BindMobile     BuiltinString  `json:"BindMobile"`
	BindUin        int            `json:"BindUin"`
	BitFlag        int            `json:"BitFlag"`
	City           string         `json:"City"`
	Country        string         `json:"Country"`
	DisturbSetting DisturbSetting `json:"DisturbSetting"`
	Experience     int            `json:"Experience"`
	FaceBookFlag   int            `json:"FaceBookFlag"`
	Fbtoken        string         `json:"Fbtoken"`
	FbuserId       int            `json:"FbuserId"`
	FbuserName     string         `json:"FbuserName"`
	GmailList      GmailList      `json:"GmailList"`
	ImgBuf         BuiltinBuffer  `json:"ImgBuf"`
	ImgLen         int            `json:"ImgLen"`
	Level          int            `json:"Level"`
	LevelHighExp   int            `json:"LevelHighExp"`
	LevelLowExp    int            `json:"LevelLowExp"`
	NickName       BuiltinString  `json:"NickName"`
	PersonalCard   int            `json:"PersonalCard"`
	PluginFlag     int            `json:"PluginFlag"`
	PluginSwitch   int            `json:"PluginSwitch"`
	Point          int            `json:"Point"`
	Province       string         `json:"Province"`
	Sex            int            `json:"Sex"`
	Signature      string         `json:"Signature"`
	Status         int            `json:"Status"`
	TxnewsCategory int            `json:"TxnewsCategory"`
	UserName       BuiltinString  `json:"UserName"`
	VerifyFlag     int            `json:"VerifyFlag"`
	VerifyInfo     string         `json:"VerifyInfo"`
	Weibo          string         `json:"Weibo"`
	WeiboFlag      int            `json:"WeiboFlag"`
	WeiboNickname  string         `json:"WeiboNickname"`
}

type UserInfoExt struct {
	BbmnickName         string              `json:"BbmnickName"`
	Bbpin               string              `json:"Bbpin"`
	Bbppid              string              `json:"Bbppid"`
	BigChatRoomInvite   int                 `json:"BigChatRoomInvite"`
	BigChatRoomQuota    int                 `json:"BigChatRoomQuota"`
	BigChatRoomSize     int                 `json:"BigChatRoomSize"`
	BigHeadImgUrl       string              `json:"BigHeadImgUrl"`
	ExtStatus           int64               `json:"ExtStatus"`
	ExtXml              BuiltinString       `json:"ExtXml"`
	F2FpushSound        string              `json:"F2FpushSound"`
	GoogleContactName   string              `json:"GoogleContactName"`
	GrayscaleFlag       int                 `json:"GrayscaleFlag"`
	IdcardNum           string              `json:"IdcardNum"`
	Kfinfo              string              `json:"Kfinfo"`
	LinkedinContactItem LinkedinContactItem `json:"LinkedinContactItem"`
	MainAcctType        int                 `json:"MainAcctType"`
	MsgPushSound        string              `json:"MsgPushSound"`
	MyBrandList         string              `json:"MyBrandList"`
	PatternLockInfo     PatternLockInfo     `json:"PatternLockInfo"`
	PaySetting          int                 `json:"PaySetting"`
	PayWalletType       int                 `json:"PayWalletType"`
	RealName            string              `json:"RealName"`
	RegCountry          string              `json:"RegCountry"`
	SafeDevice          int                 `json:"SafeDevice"`
	SafeDeviceList      SafeDeviceList      `json:"SafeDeviceList"`
	SafeMobile          string              `json:"SafeMobile"`
	SecurityDeviceId    string              `json:"SecurityDeviceId"`
	SmallHeadImgUrl     string              `json:"SmallHeadImgUrl"`
	SnsUserInfo         SnsUserInfo         `json:"SnsUserInfo"`
	UserStatus          int                 `json:"UserStatus"`
	VoipPushSound       string              `json:"VoipPushSound"`
	WalletRegion        int                 `json:"WalletRegion"`
	WeiDianInfo         string              `json:"WeiDianInfo"`
}

type PatternLockInfo struct {
	LockStatus     int  `json:"LockStatus"`
	PatternVersion int  `json:"PatternVersion"`
	Sign           Sign `json:"Sign"`
}

type Sign struct {
	Buffer string `json:"buffer"`
	ILen   int    `json:"iLen"`
}

type SafeDeviceList struct {
	Count int          `json:"Count"`
	List  []SafeDevice `json:"List"`
}

type SafeDevice struct {
	CreateTime int    `json:"CreateTime"`
	DeviceType string `json:"DeviceType"`
	Name       string `json:"Name"`
	Uuid       string `json:"Uuid"`
}

type SnsUserInfo struct {
	SnsBgimgId    string `json:"SnsBgimgId"`
	SnsBgobjectId uint64 `json:"SnsBgobjectId"`
	SnsFlag       int    `json:"SnsFlag"`
	SnsFlagEx     int    `json:"SnsFlagEx"`
}

type DisturbSetting struct {
	AllDaySetting int       `json:"AllDaySetting"`
	AllDayTim     TimeRange `json:"AllDayTim"`
	NightSetting  int       `json:"NightSetting"`
	NightTime     TimeRange `json:"NightTime"`
}

type TimeRange struct {
	BeginTime int `json:"BeginTime"`
	EndTime   int `json:"EndTime"`
}

type GmailList struct {
	Count int         `json:"Count"`
	List  []GmailAcct `json:"List"`
}

type GmailAcct struct {
	GmailAcct    string `json:"GmailAcct"`
	GmailErrCode int    `json:"GmailErrCode"`
	GmailSwitch  int    `json:"GmailSwitch"`
}

type UserProfile struct {
	BaseResponse BaseResponse `json:"baseResponse"`
	UserInfo     UserInfo     `json:"userInfo"`
	UserInfoExt  UserInfoExt  `json:"userInfoExt"`
}

type UserImg struct {
	BigHeadImgUrl   string `json:"BigHeadImgUrl"`
	ImgBuf          any    `json:"ImgBuf"`
	ImgLen          int64  `json:"ImgLen"`
	ImgMd5          string `json:"ImgMd5"`
	ImgType         int    `json:"ImgType"`
	SmallHeadImgUrl string `json:"SmallHeadImgUrl"`
}
