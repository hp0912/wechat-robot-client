package robot

type MmtlsClient struct {
	Shakehandpubkey    []byte
	Shakehandpubkeylen int32
	Shakehandprikey    []byte
	Shakehandprikeylen int32

	Shakehandpubkey_2   []byte
	Shakehandpubkeylen2 int32
	Shakehandprikey_2   []byte
	Shakehandprikeylen2 int32

	Mserverpubhashs     []byte
	ServerSeq           int
	ClientSeq           int
	ShakehandECDHkey    []byte
	ShakehandECDHkeyLen int32

	Encrptmmtlskey  []byte
	Decryptmmtlskey []byte
	EncrptmmtlsIv   []byte
	DecryptmmtlsIv  []byte

	CurDecryptSeqIv []byte
	CurEncryptSeqIv []byte

	Decrypt_part2_hash256            []byte
	Decrypt_part3_hash256            []byte
	ShakehandECDHkeyhash             []byte
	Hkdfexpand_pskaccess_key         []byte
	Hkdfexpand_pskrefresh_key        []byte
	HkdfExpand_info_serverfinish_key []byte
	Hkdfexpand_clientfinish_key      []byte
	Hkdfexpand_secret_key            []byte

	Hkdfexpand_application_key []byte
	Encrptmmtlsapplicationkey  []byte
	Decryptmmtlsapplicationkey []byte
	EncrptmmtlsapplicationIv   []byte
	DecryptmmtlsapplicationIv  []byte

	Earlydatapart       []byte
	Newsendbufferhashs  []byte
	Encrptshortmmtlskey []byte
	Encrptshortmmtlsiv  []byte
	Decrptshortmmtlskey []byte
	Decrptshortmmtlsiv  []byte

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
	Buffer []byte  `protobuf:"bytes,2,opt,name=buffer" json:"buffer,omitempty"`
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
	Ret    []byte  `json:"ret,omitempty"`
	ErrMsg []byte  `json:"errMsg,omitempty"`
}

type OplogResponse struct {
	Ret      *int32    `json:"ret,omitempty"`
	OplogRet *OplogRet `json:"oplogRet,omitempty"`
}
