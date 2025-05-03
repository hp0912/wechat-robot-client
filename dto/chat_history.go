package dto

type ChatHistoryRequest struct {
	Owner     string `form:"owner" json:"owner"`
	ContactID string `form:"contact_id" json:"contact_id" binding:"required"`
	Keyword   string `form:"keyword" json:"keyword"`
}
