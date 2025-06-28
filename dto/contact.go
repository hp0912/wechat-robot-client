package dto

type ContactListRequest struct {
	ContactIDs []string `form:"contact_ids" json:"contact_ids"`
	Type       string   `form:"type" json:"type"`
	Keyword    string   `form:"keyword" json:"keyword"`
}
