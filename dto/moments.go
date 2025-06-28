package dto

type FriendCircleGetListRequest struct {
	FristPageMd5 string `form:"frist_page_md5" json:"frist_page_md5"`
	MaxID        string `form:"max_id" json:"max_id" binding:"required"`
}

type DownFriendCircleMediaRequest struct {
	Url string `form:"url" json:"url" binding:"required"`
	Key string `form:"key" json:"key"`
}
