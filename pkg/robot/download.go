package robot

type ImageSecretXml struct {
	AesKey       string `xml:"aeskey,attr"`
	CdnMidImgUrl string `xml:"cdnmidimgurl,attr"`
	Length       int64  `xml:"length,attr"`
	Md5          string `xml:"md5,attr"`
}

type ImageMessageXml struct {
	Img ImageSecretXml `xml:"img"`
}

type DownloadImageDetail struct {
	Image string `json:"Image"`
}

type DataBuffer struct {
	Buffer string `json:"buffer,omitempty"`
	ILen   int64  `json:"iLen,omitempty"`
}

type DownloadVideoDetail struct {
	BaseResponse
	MsgId    int64      `json:"msgId"`
	NewMsgId int64      `json:"newMsgId"`
	TotalLen int64      `json:"totalLen"`
	StartPos int64      `json:"startPos"`
	Data     DataBuffer `json:"data"`
}

type VoiceSecretXml struct {
	BufID        string `xml:"bufid,attr"`
	FromUsername string `xml:"fromusername,attr"`
	AesKey       string `xml:"aeskey,attr"`
	Voiceurl     string `xml:"voiceurl,attr"`
	VoiceLength  int64  `xml:"voicelength,attr"`
	Length       int64  `xml:"length,attr"`
	VoiceMd5     string `xml:"voicemd5,attr"`
}

type VoiceMessageXml struct {
	Voicemsg VoiceSecretXml `xml:"voicemsg"`
}

type DownloadVoiceDetail struct {
	BaseResponse
	MsgId       int64      `json:"msgId"`
	Offset      int64      `json:"offset"`
	Length      int64      `json:"length"`
	VoiceLength int64      `json:"voiceLength"`
	ClientMsgId string     `json:"clientMsgId"`
	Data        DataBuffer `json:"data"`
	EndFlag     int        `json:"endFlag"`
	CancelFlag  int        `json:"cancelFlag"`
	NewMsgId    int64      `json:"newMsgId"`
}

type AppAttach struct {
	TotalLen        int64  `xml:"totallen"`
	FileExt         string `xml:"fileext"`
	AttachID        string `xml:"attachid"`
	CDNAttachURL    string `xml:"cdnattachurl"`
	CDNThumbAESKey  string `xml:"cdnthumbaeskey"`
	AESKey          string `xml:"aeskey"`
	EncryVer        string `xml:"encryver"`
	FileKey         string `xml:"filekey"`
	OverwriteMsgID  string `xml:"overwrite_newmsgid"`
	FileUploadToken string `xml:"fileuploadtoken"`
}

type FileSecretXml struct {
	AppID      string    `xml:"appid,attr"`
	SDKVer     string    `xml:"sdkver,attr"`
	Title      string    `xml:"title"`
	Des        string    `xml:"des"`
	Type       int       `xml:"type"`
	Attach     AppAttach `xml:"appattach"`
	MD5        string    `xml:"md5"`
	RecordItem string    `xml:"recorditem"`
}

type AppInfo struct {
	Version int    `xml:"version"`
	AppName string `xml:"appname"`
}

type FileMessageXml struct {
	Appmsg       FileSecretXml `xml:"appmsg"`
	FromUsername string        `xml:"fromusername"`
	Scene        int           `xml:"scene"`
	AppInfo      AppInfo       `xml:"appinfo"`
}

type DownloadFileDetail struct {
	BaseResponse
	MediaID  string     `json:"mediaId"`
	UserName string     `json:"userName"`
	TotalLen int64      `json:"totalLen"`
	StartPos int64      `json:"startPos"`
	DataLen  int64      `json:"dataLen"`
	Data     DataBuffer `json:"data"`
}
