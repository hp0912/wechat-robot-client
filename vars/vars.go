package vars

import (
	"time"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/plugin"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 微信机器人客户端监听端口
var WechatClientPort string

// 微信机器人服务端地址，仅供本地调试用
var WechatServerHost string

// 机器人实例数据库
var DB *gorm.DB

// 机器人管理后台数据库
var AdminDB *gorm.DB

// redis实例
var RedisClient *redis.Client

var MessagePlugin *plugin.MessagePlugin

// 机器人启动超时
var RobotStartTimeout time.Duration

// 机器人运行时实例
var RobotRuntime = &robot.Robot{}

// 任务调度器实例
var CronManager CronManagerInterface

// 歌曲搜索Api
// var MusicSearchApi = "https://www.hhlqilongzhu.cn/api/joox/juhe_music.php"
var MusicSearchApi = "https://api.cenguigui.cn/api/music/netease/WyY_Dg.php"

var ThirdPartyApiKey string

var WordCloudUrl string

var UploadFileChunkSize int64 = 50000

var AtAllRegexp = `@所有人(?: | )`

var TrimAtRegexp = `@[^ | ]+?(?: | )`

var OfficialAccount = map[string]bool{
	"filehelper":            true,
	"newsapp":               true,
	"fmessage":              true,
	"weibo":                 true,
	"qqmail":                true,
	"tmessage":              true,
	"qmessage":              true,
	"qqsync":                true,
	"floatbottle":           true,
	"lbsapp":                true,
	"shakeapp":              true,
	"medianote":             true,
	"qqfriend":              true,
	"readerapp":             true,
	"blogapp":               true,
	"facebookapp":           true,
	"masssendapp":           true,
	"meishiapp":             true,
	"feedsapp":              true,
	"voip":                  true,
	"blogappweixin":         true,
	"weixin":                true,
	"brandsessionholder":    true,
	"weixinreminder":        true,
	"officialaccounts":      true,
	"notification_messages": true,
	"wxitil":                true,
	"userexperience_alarm":  true,
	"exmail_tool":           true,
	"mphelper":              true,
}
