package dto

// AddKnowledgeDocumentRequest 添加知识库文档请求
type AddKnowledgeDocumentRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Source   string `json:"source"`
	Category string `json:"category"`
}

// UpdateKnowledgeDocumentRequest 更新知识库文档请求
type UpdateKnowledgeDocumentRequest struct {
	ID      int64  `json:"id" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Source  string `json:"source"`
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

// AddImageKnowledgeRequest 添加图片知识库文档请求
type AddImageKnowledgeRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url" binding:"required"`
	Category    string `json:"category"`
}

// DeleteImageKnowledgeRequest 删除图片知识库文档请求
type DeleteImageKnowledgeRequest struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// ListImageKnowledgeRequest 图片知识库列表请求
type ListImageKnowledgeRequest struct {
	Category string `form:"category"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// SearchImageKnowledgeByTextRequest 以文搜图请求
type SearchImageKnowledgeByTextRequest struct {
	Query    string `json:"query" binding:"required"`
	Category string `json:"category"`
	Limit    int    `json:"limit"`
}

// SearchImageKnowledgeByImageRequest 以图搜图请求
type SearchImageKnowledgeByImageRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
	Category string `json:"category"`
	Limit    int    `json:"limit"`
}

// CreateKnowledgeCategoryRequest 创建知识库分类请求
type CreateKnowledgeCategoryRequest struct {
	Code        string `json:"code" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=text image"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateKnowledgeCategoryRequest 更新知识库分类请求
type UpdateKnowledgeCategoryRequest struct {
	ID          int64  `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// DeleteKnowledgeCategoryRequest 删除知识库分类请求
type DeleteKnowledgeCategoryRequest struct {
	ID int64 `json:"id" binding:"required"`
}

// ListKnowledgeCategoryRequest 获取知识库分类列表请求
type ListKnowledgeCategoryRequest struct {
	Type string `form:"type" binding:"omitempty,oneof=text image"`
}
