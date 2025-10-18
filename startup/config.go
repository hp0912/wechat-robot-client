package startup

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"wechat-robot-client/vars"

	"github.com/joho/godotenv"
)

func LoadConfig() error {
	loadEnvConfig()
	return nil
}

func loadEnvConfig() {
	// 本地开发模式
	isDevMode := strings.ToLower(os.Getenv("GO_ENV")) == "dev"
	if isDevMode {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("加载本地环境变量失败，请检查是否存在 .env 文件")
		}
	}

	// 监听端口
	vars.WechatClientPort = os.Getenv("WECHAT_CLIENT_PORT")
	if vars.WechatClientPort == "" {
		log.Fatal("WECHAT_CLIENT_PORT 环境变量未设置")
	}
	vars.WechatServerHost = os.Getenv("WECHAT_SERVER_HOST")

	// mysql
	vars.MysqlSettings.Driver = os.Getenv("MYSQL_DRIVER")
	vars.MysqlSettings.Host = os.Getenv("MYSQL_HOST")
	vars.MysqlSettings.Port = os.Getenv("MYSQL_PORT")
	vars.MysqlSettings.User = os.Getenv("MYSQL_USER")
	vars.MysqlSettings.Password = os.Getenv("MYSQL_PASSWORD")
	// 机器人ID就是数据库名
	vars.MysqlSettings.Db = os.Getenv("ROBOT_CODE")
	vars.MysqlSettings.AdminDb = os.Getenv("MYSQL_ADMIN_DB")
	vars.MysqlSettings.Schema = os.Getenv("MYSQL_SCHEMA")

	// redis
	vars.RedisSettings.Host = os.Getenv("REDIS_HOST")
	vars.RedisSettings.Port = os.Getenv("REDIS_PORT")
	vars.RedisSettings.Password = os.Getenv("REDIS_PASSWORD")
	redisDb := os.Getenv("REDIS_DB")
	if redisDb == "" {
		log.Fatalf("REDIS_DB 环境变量未设置")
	} else {
		db, err := strconv.Atoi(redisDb)
		if err != nil {
			log.Fatalf("REDIS_DB 转换失败: %v", err)
		}
		vars.RedisSettings.Db = db
	}

	// rabbitmq
	vars.RabbitmqSettings.Host = os.Getenv("RABBITMQ_HOST")
	vars.RabbitmqSettings.Port = os.Getenv("RABBITMQ_PORT")
	vars.RabbitmqSettings.User = os.Getenv("RABBITMQ_USER")
	vars.RabbitmqSettings.Password = os.Getenv("RABBITMQ_PASSWORD")
	vars.RabbitmqSettings.Vhost = os.Getenv("RABBITMQ_VHOST")

	// robot
	robotStartTimeout := os.Getenv("ROBOT_START_TIMEOUT")
	if robotStartTimeout == "" {
		vars.RobotStartTimeout = 60 * time.Second
	} else {
		// 将字符串转换成int
		t, err := strconv.Atoi(robotStartTimeout)
		if err != nil {
			log.Fatalf("ROBOT_START_TIMEOUT 转换失败: %v", err)
		}
		vars.RobotStartTimeout = time.Duration(t) * time.Second
	}

	vars.ThirdPartyApiKey = os.Getenv("THIRD_PARTY_API_KEY")
	// 词云
	vars.WordCloudUrl = os.Getenv("WORD_CLOUD_URL")
	// pprof 代理地址
	vars.PprofProxyURL = os.Getenv("PPROF_PROXY_URL")
}
