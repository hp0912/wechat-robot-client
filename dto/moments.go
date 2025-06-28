package dto

type FriendCircleGetListRequest struct {
	FristPageMd5 string `json:"frist_page_md5" binding:"required"`
	MaxID        string `json:"max_id" binding:"required"`
}

type DownFriendCircleMediaRequest struct {
	Url string `json:"url" binding:"required"`
	Key string `json:"key"`
}
