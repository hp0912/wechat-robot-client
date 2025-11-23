package dto

type MessageCommonRequest struct {
	MessageID int64 `form:"message_id" json:"message_id" binding:"required"`
}

type SendMessageCommonRequest struct {
	ToWxid string `form:"to_wxid" json:"to_wxid" binding:"required"`
}

type SendTextMessageRequest struct {
	SendMessageCommonRequest
	Content string   `form:"content" json:"content" binding:"required"`
	At      []string `form:"at" json:"at"`
}

type SendLongTextMessageRequest struct {
	SendMessageCommonRequest
	Content string `form:"content" json:"content" binding:"required"`
}

type SendAppMessageRequest struct {
	SendMessageCommonRequest
	Type int    `form:"type" json:"type" binding:"required"`
	XML  string `form:"xml" json:"xml" binding:"required"`
}

type SendMusicMessageRequest struct {
	SendMessageCommonRequest
	Song string `form:"song" json:"song" binding:"required"`
}

type TextMessageItem struct {
	Nickname  string `json:"nickname"`
	Message   string `json:"message"`
	CreatedAt int64  `json:"created_at"`
}

type SendImageMessageRequest struct {
	ToWxid      string `form:"to_wxid" json:"to_wxid" binding:"required"`
	ClientImgId string `form:"client_img_id" json:"client_img_id" binding:"required"`
	FileSize    int64  `form:"file_size" json:"file_size" binding:"required"`
	ChunkIndex  int64  `form:"chunk_index" json:"chunk_index"`
	TotalChunks int64  `form:"total_chunks" json:"total_chunks" binding:"required"`
	ImageURL    string `form:"image_url" json:"image_url"` // 冗余字段
}

type SendImageMessageByRemoteURLRequest struct {
	ToWxid    string   `form:"to_wxid" json:"to_wxid" binding:"required"`
	ImageURLs []string `form:"image_urls" json:"image_urls" binding:"required"`
}

type SendVideoMessageByRemoteURLRequest struct {
	ToWxid    string   `form:"to_wxid" json:"to_wxid" binding:"required"`
	VideoURLs []string `form:"video_urls" json:"video_urls" binding:"required"`
}

type SendFileMessageRequest struct {
	ToWxid          string `form:"to_wxid" json:"to_wxid" binding:"required"`
	ClientAppDataId string `form:"client_app_data_id" json:"client_app_data_id" binding:"required"`
	Filename        string `form:"filename" json:"filename" binding:"required"`
	FileHash        string `form:"file_hash" json:"file_hash" binding:"required"`
	FileSize        int64  `form:"file_size" json:"file_size" binding:"required"`
	ChunkIndex      int64  `form:"chunk_index" json:"chunk_index"`
	TotalChunks     int64  `form:"total_chunks" json:"total_chunks" binding:"required"`
}
