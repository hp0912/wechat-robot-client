package dto

type ContactListRequest struct {
	Type    string `form:"type" json:"type"`
	Keyword string `form:"keyword" json:"keyword"`
}
