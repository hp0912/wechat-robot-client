package vars

import (
	"time"
	"wechat-robot-client/plugin"

	"gorm.io/gorm"
)

var DB *gorm.DB

// 机器人消息处理插件
var MessageHandler plugin.MessageHandler

// 机器人启动超时
var RobotStartTimeout time.Duration
