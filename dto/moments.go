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
	CommnetId uint32 `form:"CommnetId" json:"CommnetId"`
}

type MomentPrivacySettingsRequest struct {
	Function uint32 `form:"Function" json:"Function"`
	Value    uint32 `form:"Value" json:"Value"`
}
