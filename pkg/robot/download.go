package robot

import "encoding/xml"

type Section struct {
	DataLen  int64 `json:"DataLen"`
	StartPos int64 `json:"StartPos"`
}

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

type VideoSecretXml struct {
	AesKey            string `xml:"aeskey,attr"`
	CdnVideoUrl       string `xml:"cdnvideourl,attr"`
	CdnThumbAesKey    string `xml:"cdnthumbaeskey,attr"`
	CdnThumbUrl       string `xml:"cdnthumburl,attr"`
	Length            int64  `xml:"length,attr"`
	PlayLength        int64  `xml:"playlength,attr"`
	CdnThumbLength    int64  `xml:"cdnthumblength,attr"`
	CdnThumbWidth     int64  `xml:"cdnthumbwidth,attr"`
	CdnThumbHeight    int64  `xml:"cdnthumbheight,attr"`
	FromUserName      string `xml:"fromusername,attr"`
	Md5               string `xml:"md5,attr"`
	NewMd5            string `xml:"newmd5,attr"`
	IsPlaceholder     string `xml:"isplaceholder,attr"`
	RawMd5            string `xml:"rawmd5,attr"`
	RawLength         int64  `xml:"rawlength,attr"`
	CdnRawVideoUrl    string `xml:"cdnrawvideourl,attr"`
	CdnRawVideoAesKey string `xml:"cdnrawvideoaeskey,attr"`
	OverwriteNewMsgId int64  `xml:"overwritenewmsgid,attr"`
	OriginSourceMd5   string `xml:"originsourcemd5,attr"`
	IsAd              string `xml:"isad,attr"`
}

type VideoMessageXml struct {
	XMLName  xml.Name       `xml:"msg"`
	VideoMsg VideoSecretXml `xml:"videomsg"`
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

type CdnDownloadImgRequest struct {
	Wxid       string `json:"Wxid"`
	FileNo     string `json:"FileNo"`
	FileAesKey string `json:"FileAesKey"`
}

type DownloadVideoRequest struct {
	Wxid         string  `json:"Wxid"`
	MsgId        int64   `json:"MsgId"`
	CompressType int     `json:"CompressType"`
	DataLen      int64   `json:"DataLen"`
	Section      Section `json:"Section"`
	ToWxid       string  `json:"ToWxid"`
}

type DownloadVoiceRequest struct {
	Wxid         string `json:"Wxid"`
	MsgId        int64  `json:"MsgId"`
	Length       int64  `json:"Length"`
	FromUserName string `json:"FromUserName"`
	Bufid        string `json:"Bufid"`
}

type DownloadFileRequest struct {
	Wxid     string  `json:"Wxid"`
	AttachId string  `json:"AttachId"`
	AppID    string  `json:"AppID"`
	UserName string  `json:"UserName"`
	DataLen  int64   `json:"DataLen"`
	Section  Section `json:"Section"`
}
