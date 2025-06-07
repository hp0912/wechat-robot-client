package robot

type UserInfo struct {
	AlbumBgimgId   string           `json:"AlbumBgimgId"`
	AlbumFlag      int              `json:"AlbumFlag"`
	AlbumStyle     int              `json:"AlbumStyle"`
	Alias          string           `json:"Alias"`
	BindEmail      SKBuiltinStringT `json:"BindEmail"`
	BindMobile     SKBuiltinStringT `json:"BindMobile"`
	BindUin        int              `json:"BindUin"`
	BitFlag        int              `json:"BitFlag"`
	City           string           `json:"City"`
	Country        string           `json:"Country"`
	DisturbSetting DisturbSetting   `json:"DisturbSetting"`
	Experience     int              `json:"Experience"`
	FaceBookFlag   int              `json:"FaceBookFlag"`
	Fbtoken        string           `json:"Fbtoken"`
	FbuserId       int              `json:"FbuserId"`
	FbuserName     string           `json:"FbuserName"`
	GmailList      GmailList        `json:"GmailList"`
	ImgBuf         SKBuiltinBufferT `json:"ImgBuf"`
	ImgLen         int              `json:"ImgLen"`
	Level          int              `json:"Level"`
	LevelHighExp   int              `json:"LevelHighExp"`
	LevelLowExp    int              `json:"LevelLowExp"`
	NickName       SKBuiltinStringT `json:"NickName"`
	PersonalCard   int              `json:"PersonalCard"`
	PluginFlag     int              `json:"PluginFlag"`
	PluginSwitch   int              `json:"PluginSwitch"`
	Point          int              `json:"Point"`
	Province       string           `json:"Province"`
	Sex            int              `json:"Sex"`
	Signature      string           `json:"Signature"`
	Status         int              `json:"Status"`
	TxnewsCategory int              `json:"TxnewsCategory"`
	UserName       SKBuiltinStringT `json:"UserName"`
	VerifyFlag     int              `json:"VerifyFlag"`
	VerifyInfo     string           `json:"VerifyInfo"`
	Weibo          string           `json:"Weibo"`
	WeiboFlag      int              `json:"WeiboFlag"`
	WeiboNickname  string           `json:"WeiboNickname"`
}

type UserInfoExt struct {
	SnsUserInfo         *SnsUserInfo         `protobuf:"bytes,1,opt,name=SnsUserInfo" json:"SnsUserInfo,omitempty"`
	MyBrandList         *string              `protobuf:"bytes,2,opt,name=MyBrandList" json:"MyBrandList,omitempty"`
	MsgPushSound        *string              `protobuf:"bytes,3,opt,name=MsgPushSound" json:"MsgPushSound,omitempty"`
	VoipPushSound       *string              `protobuf:"bytes,4,opt,name=VoipPushSound" json:"VoipPushSound,omitempty"`
	BigChatRoomSize     *uint32              `protobuf:"varint,5,opt,name=BigChatRoomSize" json:"BigChatRoomSize,omitempty"`
	BigChatRoomQuota    *uint32              `protobuf:"varint,6,opt,name=BigChatRoomQuota" json:"BigChatRoomQuota,omitempty"`
	BigChatRoomInvite   *uint32              `protobuf:"varint,7,opt,name=BigChatRoomInvite" json:"BigChatRoomInvite,omitempty"`
	SafeMobile          *string              `protobuf:"bytes,8,opt,name=SafeMobile" json:"SafeMobile,omitempty"`
	BigHeadImgUrl       *string              `protobuf:"bytes,9,opt,name=BigHeadImgUrl" json:"BigHeadImgUrl,omitempty"`
	SmallHeadImgUrl     *string              `protobuf:"bytes,10,opt,name=SmallHeadImgUrl" json:"SmallHeadImgUrl,omitempty"`
	MainAcctType        *uint32              `protobuf:"varint,11,opt,name=MainAcctType" json:"MainAcctType,omitempty"`
	ExtXml              *SKBuiltinStringT    `protobuf:"bytes,12,opt,name=ExtXml" json:"ExtXml,omitempty"`
	SafeDeviceList      *SafeDeviceList      `protobuf:"bytes,13,opt,name=SafeDeviceList" json:"SafeDeviceList,omitempty"`
	SafeDevice          *uint32              `protobuf:"varint,14,opt,name=SafeDevice" json:"SafeDevice,omitempty"`
	GrayscaleFlag       *uint32              `protobuf:"varint,15,opt,name=GrayscaleFlag" json:"GrayscaleFlag,omitempty"`
	GoogleContactName   *string              `protobuf:"bytes,16,opt,name=GoogleContactName" json:"GoogleContactName,omitempty"`
	IdcardNum           *string              `protobuf:"bytes,17,opt,name=IdcardNum" json:"IdcardNum,omitempty"`
	RealName            *string              `protobuf:"bytes,18,opt,name=RealName" json:"RealName,omitempty"`
	RegCountry          *string              `protobuf:"bytes,19,opt,name=RegCountry" json:"RegCountry,omitempty"`
	Bbppid              *string              `protobuf:"bytes,20,opt,name=Bbppid" json:"Bbppid,omitempty"`
	Bbpin               *string              `protobuf:"bytes,21,opt,name=Bbpin" json:"Bbpin,omitempty"`
	BbmnickName         *string              `protobuf:"bytes,22,opt,name=BbmnickName" json:"BbmnickName,omitempty"`
	LinkedinContactItem *LinkedinContactItem `protobuf:"bytes,23,opt,name=LinkedinContactItem" json:"LinkedinContactItem,omitempty"`
	Kfinfo              *string              `protobuf:"bytes,24,opt,name=Kfinfo" json:"Kfinfo,omitempty"`
	PatternLockInfo     *PatternLockInfo     `protobuf:"bytes,25,opt,name=PatternLockInfo" json:"PatternLockInfo,omitempty"`
	SecurityDeviceId    *string              `protobuf:"bytes,26,opt,name=SecurityDeviceId" json:"SecurityDeviceId,omitempty"`
	PayWalletType       *uint32              `protobuf:"varint,27,opt,name=PayWalletType" json:"PayWalletType,omitempty"`
	WeiDianInfo         *string              `protobuf:"bytes,28,opt,name=WeiDianInfo" json:"WeiDianInfo,omitempty"`
	WalletRegion        *uint32              `protobuf:"varint,29,opt,name=WalletRegion" json:"WalletRegion,omitempty"`
	ExtStatus           *uint64              `protobuf:"varint,30,opt,name=ExtStatus" json:"ExtStatus,omitempty"`
	F2FpushSound        *string              `protobuf:"bytes,31,opt,name=F2FpushSound" json:"F2FpushSound,omitempty"`
	UserStatus          *uint32              `protobuf:"varint,32,opt,name=UserStatus" json:"UserStatus,omitempty"`
	PaySetting          *uint64              `protobuf:"varint,33,opt,name=PaySetting" json:"PaySetting,omitempty"`
}

type PatternLockInfo struct {
	PatternVersion *uint32           `protobuf:"varint,1,opt,name=PatternVersion" json:"PatternVersion,omitempty"`
	Sign           *SKBuiltinBufferT `protobuf:"bytes,2,opt,name=Sign" json:"Sign,omitempty"`
	LockStatus     *uint32           `protobuf:"varint,3,opt,name=LockStatus" json:"LockStatus,omitempty"`
}

type Sign struct {
	Buffer string `json:"buffer"`
	ILen   int    `json:"iLen"`
}

type SafeDeviceList struct {
	Count *int32        `protobuf:"varint,1,opt,name=Count" json:"Count,omitempty"`
	List  []*SafeDevice `protobuf:"bytes,2,rep,name=List" json:"List,omitempty"`
}

type SafeDevice struct {
	Name       *string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
	Uuid       *string `protobuf:"bytes,2,opt,name=Uuid" json:"Uuid,omitempty"`
	DeviceType *string `protobuf:"bytes,3,opt,name=DeviceType" json:"DeviceType,omitempty"`
	CreateTime *uint32 `protobuf:"varint,4,opt,name=CreateTime" json:"CreateTime,omitempty"`
}

type SnsUserInfo struct {
	SnsFlag       *uint32 `protobuf:"varint,1,opt,name=SnsFlag" json:"SnsFlag,omitempty"`
	SnsBgimgId    *string `protobuf:"bytes,2,opt,name=SnsBgimgId" json:"SnsBgimgId,omitempty"`
	SnsBgobjectId *uint64 `protobuf:"varint,3,opt,name=SnsBgobjectId" json:"SnsBgobjectId,omitempty"`
	SnsFlagEx     *uint32 `protobuf:"varint,4,opt,name=SnsFlagEx" json:"SnsFlagEx,omitempty"`
}

type DisturbTimeSpan struct {
	BeginTime *uint32 `protobuf:"varint,1,opt,name=BeginTime" json:"BeginTime,omitempty"`
	EndTime   *uint32 `protobuf:"varint,2,opt,name=EndTime" json:"EndTime,omitempty"`
}

type DisturbSetting struct {
	NightSetting  *uint32          `protobuf:"varint,1,opt,name=NightSetting" json:"NightSetting,omitempty"`
	NightTime     *DisturbTimeSpan `protobuf:"bytes,2,opt,name=NightTime" json:"NightTime,omitempty"`
	AllDaySetting *uint32          `protobuf:"varint,3,opt,name=AllDaySetting" json:"AllDaySetting,omitempty"`
	AllDayTim     *DisturbTimeSpan `protobuf:"bytes,4,opt,name=AllDayTim" json:"AllDayTim,omitempty"`
}

type TimeRange struct {
	BeginTime int `json:"BeginTime"`
	EndTime   int `json:"EndTime"`
}

type GmailInfo struct {
	GmailAcct    *string `protobuf:"bytes,1,opt,name=GmailAcct" json:"GmailAcct,omitempty"`
	GmailSwitch  *uint32 `protobuf:"varint,2,opt,name=GmailSwitch" json:"GmailSwitch,omitempty"`
	GmailErrCode *uint32 `protobuf:"varint,3,opt,name=GmailErrCode" json:"GmailErrCode,omitempty"`
}

type GmailList struct {
	Count *uint32      `protobuf:"varint,1,opt,name=Count" json:"Count,omitempty"`
	List  []*GmailInfo `protobuf:"bytes,2,rep,name=List" json:"List,omitempty"`
}

type GetProfileResponse struct {
	BaseResponse *BaseResponse `json:"baseResponse,omitempty"`
	UserInfo     *ModUserInfo  `json:"userInfo,omitempty"`
	UserInfoExt  *UserInfoExt  `json:"userInfoExt,omitempty"`
}

type UserImg struct {
	BigHeadImgUrl   string `json:"BigHeadImgUrl"`
	ImgBuf          any    `json:"ImgBuf"`
	ImgLen          int64  `json:"ImgLen"`
	ImgMd5          string `json:"ImgMd5"`
	ImgType         int    `json:"ImgType"`
	SmallHeadImgUrl string `json:"SmallHeadImgUrl"`
}

type ModUserInfo struct {
	BitFlag        *uint32           `protobuf:"varint,1,opt,name=BitFlag" json:"BitFlag,omitempty"`
	UserName       *SKBuiltinStringT `protobuf:"bytes,2,opt,name=UserName" json:"UserName,omitempty"`
	NickName       *SKBuiltinStringT `protobuf:"bytes,3,opt,name=NickName" json:"NickName,omitempty"`
	BindUin        *uint32           `protobuf:"varint,4,opt,name=BindUin" json:"BindUin,omitempty"`
	BindEmail      *SKBuiltinStringT `protobuf:"bytes,5,opt,name=BindEmail" json:"BindEmail,omitempty"`
	BindMobile     *SKBuiltinStringT `protobuf:"bytes,6,opt,name=BindMobile" json:"BindMobile,omitempty"`
	Status         *uint32           `protobuf:"varint,7,opt,name=Status" json:"Status,omitempty"`
	ImgLen         *uint32           `protobuf:"varint,8,opt,name=ImgLen" json:"ImgLen,omitempty"`
	ImgBuf         []byte            `protobuf:"bytes,9,opt,name=ImgBuf" json:"ImgBuf,omitempty"`
	Sex            *int32            `protobuf:"varint,10,opt,name=Sex" json:"Sex,omitempty"`
	Province       *string           `protobuf:"bytes,11,opt,name=Province" json:"Province,omitempty"`
	City           *string           `protobuf:"bytes,12,opt,name=City" json:"City,omitempty"`
	Signature      *string           `protobuf:"bytes,13,opt,name=Signature" json:"Signature,omitempty"`
	PersonalCard   *uint32           `protobuf:"varint,14,opt,name=PersonalCard" json:"PersonalCard,omitempty"`
	DisturbSetting *DisturbSetting   `protobuf:"bytes,15,opt,name=DisturbSetting" json:"DisturbSetting,omitempty"`
	PluginFlag     *uint32           `protobuf:"varint,16,opt,name=PluginFlag" json:"PluginFlag,omitempty"`
	VerifyFlag     *uint32           `protobuf:"varint,17,opt,name=VerifyFlag" json:"VerifyFlag,omitempty"`
	VerifyInfo     *string           `protobuf:"bytes,18,opt,name=VerifyInfo" json:"VerifyInfo,omitempty"`
	Point          *uint32           `protobuf:"varint,19,opt,name=Point" json:"Point,omitempty"`
	Experience     *uint32           `protobuf:"varint,20,opt,name=Experience" json:"Experience,omitempty"`
	Level          *uint32           `protobuf:"varint,21,opt,name=Level" json:"Level,omitempty"`
	LevelLowExp    *uint32           `protobuf:"varint,22,opt,name=LevelLowExp" json:"LevelLowExp,omitempty"`
	LevelHighExp   *uint32           `protobuf:"varint,23,opt,name=LevelHighExp" json:"LevelHighExp,omitempty"`
	Weibo          *string           `protobuf:"bytes,24,opt,name=Weibo" json:"Weibo,omitempty"`
	PluginSwitch   *uint32           `protobuf:"varint,25,opt,name=PluginSwitch" json:"PluginSwitch,omitempty"`
	GmailList      *GmailList        `protobuf:"bytes,26,opt,name=GmailList" json:"GmailList,omitempty"`
	Alias          *string           `protobuf:"bytes,27,opt,name=Alias" json:"Alias,omitempty"`
	WeiboNickname  *string           `protobuf:"bytes,28,opt,name=WeiboNickname" json:"WeiboNickname,omitempty"`
	WeiboFlag      *uint32           `protobuf:"varint,29,opt,name=WeiboFlag" json:"WeiboFlag,omitempty"`
	FaceBookFlag   *uint32           `protobuf:"varint,30,opt,name=FaceBookFlag" json:"FaceBookFlag,omitempty"`
	FbuserId       *uint64           `protobuf:"varint,31,opt,name=FbuserId" json:"FbuserId,omitempty"`
	FbuserName     *string           `protobuf:"bytes,32,opt,name=FbuserName" json:"FbuserName,omitempty"`
	AlbumStyle     *int32            `protobuf:"varint,33,opt,name=AlbumStyle" json:"AlbumStyle,omitempty"`
	AlbumFlag      *int32            `protobuf:"varint,34,opt,name=AlbumFlag" json:"AlbumFlag,omitempty"`
	AlbumBgimgId   *string           `protobuf:"bytes,35,opt,name=AlbumBgimgId" json:"AlbumBgimgId,omitempty"`
	TxnewsCategory *uint32           `protobuf:"varint,36,opt,name=TxnewsCategory" json:"TxnewsCategory,omitempty"`
	Fbtoken        *string           `protobuf:"bytes,37,opt,name=Fbtoken" json:"Fbtoken,omitempty"`
	Country        *string           `protobuf:"bytes,38,opt,name=Country" json:"Country,omitempty"`
}
