package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type ChatRoomMember struct {
	Base[model.ChatRoomMember]
}

func NewChatRoomMemberRepo(ctx context.Context, db *gorm.DB) *ChatRoomMember {
	return &ChatRoomMember{
		Base[model.ChatRoomMember]{
			Ctx: ctx,
			DB:  db,
		}}
}
