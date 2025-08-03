package robot

type MmtlsClient struct {
	Shakehandpubkey    string
	Shakehandpubkeylen int32
	Shakehandprikey    string
	Shakehandprikeylen int32

	Shakehandpubkey_2   string
	Shakehandpubkeylen2 int32
	Shakehandprikey_2   string
	Shakehandprikeylen2 int32

	Mserverpubhashs     string
	ServerSeq           int
	ClientSeq           int
	ShakehandECDHkey    string
	ShakehandECDHkeyLen int32

	Encrptmmtlskey  string
	Decryptmmtlskey string
	EncrptmmtlsIv   string
	DecryptmmtlsIv  string

	CurDecryptSeqIv string
	CurEncryptSeqIv string

	Decrypt_part2_hash256            string
	Decrypt_part3_hash256            string
	ShakehandECDHkeyhash             string
	Hkdfexpand_pskaccess_key         string
	Hkdfexpand_pskrefresh_key        string
	HkdfExpand_info_serverfinish_key string
	Hkdfexpand_clientfinish_key      string
	Hkdfexpand_secret_key            string

	Hkdfexpand_application_key string
	Encrptmmtlsapplicationkey  string
	Decryptmmtlsapplicationkey string
	EncrptmmtlsapplicationIv   string
	DecryptmmtlsapplicationIv  string

	Earlydatapart       string
	Newsendbufferhashs  string
	Encrptshortmmtlskey string
	Encrptshortmmtlsiv  string
	Decrptshortmmtlskey string
	Decrptshortmmtlsiv  string

	//http才需要
	Pskkey    string
	Pskiv     string
	MmtlsMode uint
}

type SKBuiltinStringT struct {
	String *string `json:"string,omitempty"`
}

type SKBuiltinBufferT struct {
	ILen   *uint32 `protobuf:"varint,1,opt,name=iLen" json:"iLen,omitempty"`
	Buffer string  `protobuf:"bytes,2,opt,name=buffer" json:"buffer,omitempty"`
}

type SKBuiltinString_S struct {
	ILen   *uint32 `protobuf:"varint,1,opt,name=iLen" json:"iLen,omitempty"`
	Buffer *string `protobuf:"bytes,2,opt,name=buffer" json:"buffer,omitempty"`
}

type CommonRequest struct {
	Wxid string `json:"wxid"`
}

type ProxyInfo struct {
	ProxyIp       string `json:"ProxyIp"`
	ProxyUser     string `json:"ProxyUser"`
	ProxyPassword string `json:"ProxyPassword"`
}

type Dns struct {
	Ip   string
	Host string
}

type Timespec struct {
	Tvsec  uint64 `json:"tv_sec"`
	Tvnsec uint64 `json:"tv_nsec"`
}

type Statfs struct {
	Type        uint64 `json:"type"`        // f_type = 26
	Fstypename  string `json:"fstypename"`  //apfs
	Flags       uint64 `json:"flags"`       //statfs f_flags =  1417728009
	Mntonname   string `json:"mntonname"`   // /
	Mntfromname string `json:"mntfromname"` // com.apple.os.update-{%{96}s}@/dev/disk0s1s1
	Fsid        uint64 `json:"fsid"`        // f_fsid[0]
}

type Stat struct {
	Inode   uint64   `json:"inode"`
	Statime Timespec `json:"st_atime"`
	Stmtime Timespec `json:"st_mtime"`
	Stctime Timespec `json:"st_ctime"`
	Stbtime Timespec `json:"st_btime"`
}

// BaseResponse 大部分返回对象都携带该信息
type BaseResponse struct {
	Ret    int               `json:"ret"`
	ErrMsg *SKBuiltinStringT `json:"errMsg"`
}

func (b BaseResponse) Ok() bool {
	return b.Ret == 0
}

type OplogRet struct {
	Count  *uint32 `json:"count,omitempty"`
	Ret    string  `json:"ret,omitempty"`
	ErrMsg string  `json:"errMsg,omitempty"`
}

type OplogResponse struct {
	Ret      *int32    `json:"ret,omitempty"`
	OplogRet *OplogRet `json:"oplogRet,omitempty"`
}
