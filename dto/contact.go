package dto

type ContactListRequest struct {
	ContactIDs []string `form:"contact_ids" json:"contact_ids"`
	Type       string   `form:"type" json:"type"`
	Keyword    string   `form:"keyword" json:"keyword"`
}

type FriendPassVerifyRequest struct {
	SystemMessageID int64 `form:"system_message_id" json:"system_message_id"`
}

type FriendDeleteRequest struct {
	ContactID string `form:"contact_id" json:"contact_id"`
}
