package dto

type ContactListRequest struct {
	Owner   string `form:"owner" json:"owner"`
	Type    string `form:"type" json:"type"`
	Keyword string `form:"keyword" json:"keyword"`
}
