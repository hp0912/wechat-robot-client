package dto

type FriendSettingsRequest struct {
	ContactID string `form:"contact_id" json:"contact_id" binding:"required"`
}
