package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type Message struct {
	Base[model.Message]
}

func NewMessageRepo(ctx context.Context, db *gorm.DB) *Message {
	return &Message{
		Base[model.Message]{
			Ctx: ctx,
			DB:  db,
		}}
}
