package vars

import (
	"time"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/plugin"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 机器人实例数据库
var DB *gorm.DB

// 机器人管理后台数据库
var AdminDB *gorm.DB

// redis实例
var RedisClient *redis.Client

// 机器人消息处理插件
var MessageHandler plugin.MessageHandler

// 机器人启动超时
var RobotStartTimeout time.Duration

// 机器人运行时实例
var RobotRuntime = &robot.Robot{}

// 任务调度器实例
var CronManager CronManagerInterface

// 歌曲搜索Api
var MusicSearchApi = "https://www.hhlqilongzhu.cn/api/joox/juhe_music.php"

var WordCloudUrl string
