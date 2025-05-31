package dto

type WordCloudRequest struct {
	ChatRoomID string `json:"chat_room_id"`
	Content    string `json:"content"`
	Mode       string `json:"mode"`
}
