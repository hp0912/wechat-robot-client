package robot

import "encoding/xml"

type TimelineObject struct {
	XMLName                xml.Name      `xml:"TimelineObject"`
	ID                     string        `xml:"id"`
	Username               string        `xml:"username"`
	CreateTime             string        `xml:"createTime"`
	ContentDesc            string        `xml:"contentDesc"`
	ContentDescShowType    int           `xml:"contentDescShowType"`
	ContentDescScene       int           `xml:"contentDescScene"`
	Private                string        `xml:"private"`
	SightFolded            int           `xml:"sightFolded,omitempty"`
	ShowFlag               int           `xml:"showFlag,omitempty"`
	ContentAttr            string        `xml:"contentattr,omitempty"`
	SourceUserName         string        `xml:"sourceUserName"`
	SourceNickName         string        `xml:"sourceNickName"`
	PublicUserName         string        `xml:"publicUserName"`
	PublicBrandContactType int           `xml:"publicBrandContactType,omitempty"`
	StatisticsData         string        `xml:"statisticsData"`
	StatExtStr             string        `xml:"statExtStr,omitempty"`
	CanvasInfoXML          string        `xml:"canvasInfoXml,omitempty"`
	AppInfo                AppInfo       `xml:"appInfo"`
	WeappInfo              WeappInfo     `xml:"weappInfo,omitempty"`
	ContentObject          ContentObject `xml:"ContentObject"`
	ActionInfo             ActionInfo    `xml:"actionInfo"`
	Location               Location      `xml:"location"`
	StreamVideo            StreamVideo   `xml:"streamvideo"`
}

type WeappInfo struct {
	AppUserName      string `xml:"appUserName"`
	PagePath         string `xml:"pagePath"`
	Version          string `xml:"version"`
	IsHidden         int    `xml:"isHidden"`
	DebugMode        string `xml:"debugMode"`
	ShareActionId    string `xml:"shareActionId"`
	IsGame           string `xml:"isGame"`
	MessageExtraData string `xml:"messageExtraData"`
	SubType          string `xml:"subType"`
	PreloadResources string `xml:"preloadResources"`
}

type ContentObject struct {
	ContentStyle    string    `xml:"contentStyle"`
	ContentSubStyle string    `xml:"contentSubStyle,omitempty"`
	Title           string    `xml:"title"`
	Description     string    `xml:"description"`
	ContentUrl      string    `xml:"contentUrl"`
	MediaList       MediaList `xml:"mediaList"`
}

type MediaList struct {
	Media []Media `xml:"media"`
}

type Media struct {
	ID              string          `xml:"id"`
	Type            string          `xml:"type"`
	Title           string          `xml:"title"`
	Description     string          `xml:"description"`
	Private         string          `xml:"private"`
	UserData        string          `xml:"userData,omitempty"`
	SubType         string          `xml:"subType,omitempty"`
	VideoSize       VideoSize       `xml:"videoSize,omitempty"`
	HD              URL             `xml:"hd"`
	UHD             URL             `xml:"uhd"`
	URL             URL             `xml:"url"`
	Thumb           Thumb           `xml:"thumb"`
	Size            Size            `xml:"size"`
	VideoDuration   string          `xml:"videoDuration,omitempty"`
	VideoColdDLRule VideoColdDLRule `xml:"VideoColdDLRule,omitempty"`
}

type VideoSize struct {
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
}

type URL struct {
	Type     string `xml:"type,attr"`
	MD5      string `xml:"md5,attr"`
	VideoMD5 string `xml:"videomd5,attr"`
	Value    string `xml:",chardata"`
}

type Thumb struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type Size struct {
	Width     string `xml:"width,attr"`
	Height    string `xml:"height,attr"`
	TotalSize string `xml:"totalSize,attr"`
}

type VideoColdDLRule struct {
	All string `xml:"All"`
}

type ActionInfo struct {
	AppMsg AppMsg `xml:"appMsg"`
}

type AppMsg struct {
	MediaTagName  string `xml:"mediaTagName,omitempty"`
	MessageExt    string `xml:"messageExt,omitempty"`
	MessageAction string `xml:"messageAction"`
}

type Location struct {
	PoiClassifyId   string `xml:"poiClassifyId,attr"`
	PoiName         string `xml:"poiName,attr"`
	PoiAddress      string `xml:"poiAddress,attr"`
	PoiClassifyType string `xml:"poiClassifyType,attr"`
	City            string `xml:"city,attr"`
}

type StreamVideo struct {
	StreamVideoUrl      string `xml:"streamvideourl"`
	StreamVideoThumbUrl string `xml:"streamvideothumburl"`
	StreamVideoWebUrl   string `xml:"streamvideoweburl"`
}

type GetListRequest struct {
	Wxid         string `json:"Wxid"`
	Fristpagemd5 string `json:"Fristpagemd5"`
	Maxid        uint64 `json:"Maxid"`
}

type GetListResponse struct {
	BaseResponse          *BaseResponse      `json:"BaseResponse,omitempty"`
	FirstPageMd5          *string            `json:"FirstPageMd5,omitempty"`
	ObjectCount           *uint32            `json:"ObjectCount,omitempty"`
	ObjectList            []*SnsObject       `json:"ObjectList,omitempty"`
	NewRequestTime        *uint32            `json:"NewRequestTime,omitempty"`
	ObjectCountForSameMd5 *uint32            `json:"ObjectCountForSameMd5,omitempty"`
	ControlFlag           *uint32            `json:"ControlFlag,omitempty"`
	ServerConfig          *SnsServerConfig   `json:"ServerConfig,omitempty"`
	AdvertiseCount        *uint32            `json:"AdvertiseCount,omitempty"`
	AdvertiseList         *string            `json:"AdvertiseList,omitempty"`
	Session               *SKBuiltinString_S `json:"Session,omitempty"`
	RecCount              *uint32            `json:"RecCount,omitempty"`
	RecList               *uint32            `json:"RecList,omitempty"`
}

type SnsObject struct {
	Id                   *uint64              `json:"Id,omitempty"`
	Username             *string              `json:"Username,omitempty"`
	Nickname             *string              `json:"Nickname,omitempty"`
	CreateTime           *uint32              `json:"CreateTime,omitempty"`
	ObjectDesc           *SKBuiltinString_S   `json:"ObjectDesc,omitempty"`
	TimelineObject       *TimelineObject      `json:"TimelineObject,omitempty"`
	LikeFlag             *uint32              `json:"LikeFlag,omitempty"`
	LikeCount            *uint32              `json:"LikeCount,omitempty"`
	LikeUserListCount    *uint32              `json:"LikeUserListCount,omitempty"`
	LikeUserList         []*SnsCommentInfo    `json:"LikeUserList,omitempty"`
	CommentCount         *uint32              `json:"CommentCount,omitempty"`
	CommentUserListCount *uint32              `json:"CommentUserListCount,omitempty"`
	CommentUserList      []*SnsCommentInfo    `json:"CommentUserList,omitempty"`
	WithUserCount        *uint32              `json:"WithUserCount,omitempty"`
	WithUserListCount    *uint32              `json:"WithUserListCount,omitempty"`
	WithUserList         []*SnsCommentInfo    `json:"WithUserList,omitempty"`
	ExtFlag              *uint32              `json:"ExtFlag,omitempty"`
	NoChange             *uint32              `json:"NoChange,omitempty"`
	GroupCount           *uint32              `json:"GroupCount,omitempty"`
	GroupList            []*SnsGroup          `json:"GroupList,omitempty"`
	IsNotRichText        *uint32              `json:"IsNotRichText,omitempty"`
	ReferUsername        *string              `json:"ReferUsername,omitempty"`
	ReferId              *uint64              `json:"ReferId,omitempty"`
	BlackListCount       *uint32              `json:"BlackListCount,omitempty"`
	BlackList            []*SKBuiltinString_S `json:"BlackList,omitempty"`
	DeleteFlag           *uint32              `json:"DeleteFlag,omitempty"`
	GroupUserCount       *uint32              `json:"GroupUserCount,omitempty"`
	GroupUser            []*SKBuiltinString_S `json:"GroupUser,omitempty"`
	ObjectOperations     []*SKBuiltinString_S `json:"ObjectOperations,omitempty"`
	SnsRedEnvelops       *SnsRedEnvelops      `json:"SnsRedEnvelops,omitempty"`
	PreDownloadInfo      *PreDownloadInfo     `json:"PreDownloadInfo,omitempty"`
	WeAppInfo            *SnsWeAppInfo        `json:"WeAppInfo,omitempty"`
}

type SnsServerConfig struct {
	PostMentionLimit      *int32 `json:"PostMentionLimit,omitempty"`
	CopyAndPasteWordLimit *int32 `json:"CopyAndPasteWordLimit,omitempty"`
}

type SnsCommentInfo struct {
	Username        *string `json:"Username,omitempty"`
	Nickname        *string `json:"Nickname,omitempty"`
	Source          *uint32 `json:"Source,omitempty"`
	Type            *uint32 `json:"Type,omitempty"`
	Content         *string `json:"Content,omitempty"`
	CreateTime      *uint32 `json:"CreateTime,omitempty"`
	CommentId       *int32  `json:"CommentId,omitempty"`
	ReplyCommentId  *int32  `json:"ReplyCommentId,omitempty"`
	ReplyUsername   *string `json:"ReplyUsername,omitempty"`
	IsNotRichText   *uint32 `json:"IsNotRichText,omitempty"`
	ReplyCommentId2 *uint64 `json:"ReplyCommentId2,omitempty"`
	CommentId2      *uint64 `json:"CommentId2,omitempty"`
	DeleteFlag      *uint32 `json:"DeleteFlag,omitempty"`
	CommentFlag     *uint32 `json:"CommentFlag,omitempty"`
}

type SnsGroup struct {
	GroupId *uint64 `json:"GroupId,omitempty"`
}

type SnsRedEnvelops struct {
	RewardCount    *uint32 `json:"RewardCount,omitempty"`
	RewardUserList *string `json:"RewardUserList,omitempty"`
	ReportId       *uint32 `json:"ReportId,omitempty"`
	ReportKey      *uint32 `json:"ReportKey,omitempty"`
	ResourceId     *uint32 `json:"ResourceId,omitempty"`
}

type PreDownloadInfo struct {
	PreDownloadPercent *uint32 `json:"PreDownloadPercent,omitempty"`
	PreDownloadNetType *uint32 `json:"PreDownloadNetType,omitempty"`
	NoPreDownloadRange *string `json:"NoPreDownloadRange,omitempty"`
}

type SnsWeAppInfo struct {
	MapPoiId    *string `json:"MapPoiId,omitempty"`
	AppId       *uint32 `json:"AppId,omitempty"`
	UserName    *string `json:"UserName,omitempty"`
	RedirectUrl *string `json:"RedirectUrl,omitempty"`
	ShowType    *uint32 `json:"ShowType,omitempty"`
	RScore      *uint32 `json:"RScore,omitempty"`
}

type DownFriendCircleMediaRequest struct {
	Wxid string `json:"Wxid"`
	Url  string `json:"Url"`
	Key  string `json:"Key"`
}
