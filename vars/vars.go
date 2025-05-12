package vars

import (
	"time"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/plugin"

	"gorm.io/gorm"
)

// 机器人实例数据库
var DB *gorm.DB

// 机器人管理后台数据库
var AdminDB *gorm.DB

// 机器人消息处理插件
var MessageHandler plugin.MessageHandler

// 机器人启动超时
var RobotStartTimeout time.Duration

// 机器人运行时实例
var RobotRuntime = &robot.Robot{}

// 歌曲搜索Api
var MusicSearchApi = "https://www.hhlqilongzhu.cn/api/dg_wyymusic.php"
