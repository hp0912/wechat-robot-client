package dto

import "wechat-robot-client/pkg/robot"

type FriendCircleGetListRequest struct {
	FristPageMd5 string `form:"frist_page_md5" json:"frist_page_md5"`
	MaxID        string `form:"max_id" json:"max_id" binding:"required"`
}

type DownFriendCircleMediaRequest struct {
	Url string `form:"url" json:"url" binding:"required"`
	Key string `form:"key" json:"key"`
}

type MomentPostRequest struct {
	Content      string                             `form:"content" json:"content"`
	MediaList    []robot.FriendCircleUploadResponse `form:"media_list" json:"media_list"`
	WithUserList []string                           `form:"with_user_list" json:"with_user_list"`
	ShareType    string                             `form:"share_type" json:"share_type" binding:"required"`
	ShareWith    []string                           `form:"share_with" json:"share_with"`
	DoNotShare   []string                           `form:"donot_share" json:"donot_share"`
}

type MomentOpRequest struct {
	Id        string `form:"Id" json:"Id" binding:"required"`
	Type      uint32 `form:"Type" json:"Type" binding:"required"`
	CommentId uint32 `form:"CommentId" json:"CommentId"`
}

type MomentPrivacySettingsRequest struct {
	Function uint32 `form:"Function" json:"Function"`
	Value    uint32 `form:"Value" json:"Value"`
}

type FriendCircleCommentRequest struct {
	Type           uint32 `form:"Type" json:"Type" binding:"required"`
	Id             string `form:"Id" json:"Id" binding:"required"`
	ReplyCommnetId uint32 `form:"ReplyCommnetId" json:"ReplyCommnetId"`
	Content        string `form:"Content" json:"Content"`
}

type FriendCircleGetDetailRequest struct {
	Towxid       string `form:"Towxid" json:"Towxid" binding:"required"`
	Fristpagemd5 string `form:"Fristpagemd5" json:"Fristpagemd5"`
	Maxid        uint64 `form:"Maxid" json:"Maxid"`
}

type FriendCircleGetIdDetailRequest struct {
	Towxid string `form:"Towxid" json:"Towxid" binding:"required"`
	Id     uint64 `form:"Id" json:"Id" binding:"required"`
}
