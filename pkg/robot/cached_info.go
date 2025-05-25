package robot

type TrustSoftData struct {
	SoftConfig *string `json:"softConfig,omitempty"`
	SoftData   []byte  `json:"softData,omitempty"`
}

type TrustResponseData struct {
	SoftData    *TrustSoftData `json:"softData,omitempty"`
	DeviceToken *string        `json:"deviceToken,omitempty"`
	TimeStamp   *uint32        `json:"timeStamp,omitempty"`
}

type TrustResponse struct {
	BaseResponse      *BaseResponse      `json:"BaseResponse,omitempty"`
	TrustResponseData *TrustResponseData `json:"TrustResponseData,omitempty"`
}

type LoginDataInfo struct {
	Type     byte
	UserName string
	PassWord string
	//伪密码
	NewPassWord string
	//登录数据 62/A16
	LoginData string
	Ticket    string
	NewType   int
	Language  string
}

type DeviceInfo struct {
	UUIDOne            string `json:"uuidone"`
	UUIDTwo            string `json:"uuidtwo"`
	Imei               string `json:"imei"`
	DeviceID           string `json:"deviceid"`
	DeviceName         string `json:"devicename"`
	DeviceMac          string `json:"Devicemac"`
	TimeZone           string `json:"timezone"`
	Language           string `json:"language"`
	DeviceBrand        string `json:"devicebrand"`
	RealCountry        string `json:"realcountry"`
	IphoneVer          string `json:"iphonever"`
	BundleID           string `json:"boudleid"`
	OsType             string `json:"ostype"`
	AdSource           string `json:"adsource"`
	OsTypeNumber       string `json:"ostypenumber"`
	CoreCount          uint32 `json:"corecount"`
	CarrierName        string `json:"carriername"`
	SoftTypeXML        string `json:"softtypexml"`
	ClientCheckDataXML string `json:"clientcheckdataxml"`
	// extInfo
	GUID1 string `json:"guid1"` //data   path  uuid
	GUID2 string `json:"guid2"` //bundle path  uuid
	Sdi   string `json:"sdi"`   //md5(uuid)

	InstallTime  uint64 `json:"installtime"`  //random
	KernBootTime uint64 `json:"kernboottime"` //random

	Sysverplist *Stat `json:"sysverplist"` ////System/Library/CoreServices/SystemVersion.plist
	Dyldcache   *Stat `json:"dyldcache"`   ///System/Library/Caches/com.apple.dyld/dyld_shared_cache_arm64
	Var         *Stat `json:"var"`         ///private/var
	Etcgroup    *Stat `json:"etcgroup"`    ///private/etc/group
	Etchosts    *Stat `json:"etchosts"`    ///private/etc/hosts

	Apfs *Statfs `json:"apfs"` //apfs

	DeviceToken TrustResponse
}

type AndroidDeviceInfo struct {
	Imei                string
	AndriodId           string
	PhoneSerial         string
	WidevineDeviceID    string
	WidevineProvisionID string
	AndriodFsId         string
	AndriodBssId        string
	AndriodSsId         string
	WLanAddress         string
	PackageSign         string
	Androidversion      string
	RadioVersion        string
	Manufacturer        string
	BuildID             string
	BuildFP             string
	BuildBoard          string
	PhoneModel          string
	Hardware            string
	Features            string
	WifiName            string
	WifiFullName        string
	KernelReleaseNumber string
	Arch                string
	SfMD5               string
	SfArmMD5            string
	SfArm64MD5          string
	SbMD5               string
}

type LoginData struct {
	Uin                        uint32
	Wxid                       string
	Pwd                        string
	Uuid                       string
	Aeskey                     []byte
	NotifyKey                  []byte
	Deviceid_str               string
	Deviceid_byte              []byte
	DeviceType                 string
	ClientVersion              int32
	DeviceName                 string
	NickName                   string
	HeadUrl                    string
	Email                      string
	Alais                      string
	Mobile                     string
	Mmtlsip                    string
	ShortHost                  string
	LongHost                   string
	Sessionkey                 []byte
	Sessionkey_2               []byte
	Autoauthkey                []byte
	Autoauthkeylen             int32
	Clientsessionkey           []byte
	Serversessionkey           []byte
	HybridEcdhPrivkey          []byte
	HybridEcdhPubkey           []byte
	HybridEcdhInitServerPubKey []byte
	Loginecdhkey               []byte
	Cooike                     []byte
	LoginMode                  string
	Proxy                      ProxyInfo
	MmtlsKey                   MmtlsClient
	DeviceToken                TrustResponse
	SyncKey                    []byte
	Data62                     string
	RomModel                   string
	Imei                       string
	SoftType                   string
	OsVersion                  string
	RsaPublicKey               []byte
	RsaPrivateKey              []byte
	Dns                        []Dns
	// 登录的Rsa 密钥版本
	LoginRsaVer uint32
	// 是否开启服务
	EnableService bool
	EcPublicKey   []byte `json:"ecpukey"`
	EcPrivateKey  []byte `json:"ecprkey"`
	Ticket        string
	LoginDataInfo LoginDataInfo
	// 设备信息62
	DeviceInfo *DeviceInfo
	//A16信息
	DeviceInfoA16 *AndroidDeviceInfo
	// 登录时间
	LoginDate int64
	// 刷新 tonken 时间
	RefreshTokenDate int64
}
