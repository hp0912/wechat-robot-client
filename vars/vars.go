package vars

import (
	"wechat-robot-client/plugin"

	"gorm.io/gorm"
)

var DB *gorm.DB
var MessageHandler plugin.MessageHandler
