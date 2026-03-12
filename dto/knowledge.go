package dto

// AddKnowledgeDocumentRequest 添加知识库文档请求
type AddKnowledgeDocumentRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Source   string `json:"source"`
	Category string `json:"category"`
}

// SearchKnowledgeRequest 搜索知识库请求
type SearchKnowledgeRequest struct {
	Query    string `json:"query" binding:"required"`
	Category string `json:"category"`
	Limit    int    `json:"limit"`
}

// ListKnowledgeRequest 列表请求
type ListKnowledgeRequest struct {
	Category string `form:"category"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// DeleteKnowledgeRequest 删除知识库文档请求
type DeleteKnowledgeRequest struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// SaveMemoryRequest 手动保存记忆请求
type SaveMemoryRequest struct {
	ContactWxID string `json:"contact_wxid" binding:"required"`
	ChatRoomID  string `json:"chat_room_id"`
	Type        string `json:"type" binding:"required"`
	Key         string `json:"key" binding:"required"`
	Content     string `json:"content" binding:"required"`
	Importance  int    `json:"importance"`
}

// SearchMemoryRequest 搜索记忆请求
type SearchMemoryRequest struct {
	ContactWxID string `form:"contact_wxid"`
	Query       string `form:"query"`
	Limit       int    `form:"limit"`
}

// DeleteMemoryRequest 删除记忆请求
type DeleteMemoryRequest struct {
	ID int64 `json:"id" binding:"required"`
}
