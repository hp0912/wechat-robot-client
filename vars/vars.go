package vars

import (
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"wechat-robot-client/interface/ai"
	"wechat-robot-client/pkg/qdrantx"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/plugin"
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

var MCPService ai.MCPService

var SkillsDir string
var SkillService ai.SkillService

// Qdrant 客户端
var QdrantClient *qdrantx.QdrantClient

// RAG & 记忆相关服务
var MemoryService ai.MemoryService
var RAGService ai.RAGService
var KnowledgeService ai.KnowledgeService
var ImageKnowledgeService ai.ImageKnowledgeService

var Webhook struct {
	URL     string
	Headers map[string]any
}

// 任务调度器实例
var CronManager CronManagerInterface

// 歌曲搜索Api
var MusicSearchApi = "https://api.cenguigui.cn/api/music/netease/WyY_Dg.php"

var SliderAccessKey string

var ThirdPartyApiKey string

var WordCloudUrl string

// Pprof 代理目标地址
var PprofProxyURL string

var AIEnded = "ended"

var UploadImageChunkSize int64 = 200 * 1000 // 200KB
var UploadFileChunkSize int64 = 200 * 1000  // 200KB

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
