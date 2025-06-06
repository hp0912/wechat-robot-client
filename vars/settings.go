package vars

type MysqlSettingS struct {
	Driver   string // 使用的数据库驱动，支持 mysql、postgres
	Host     string
	Port     string
	User     string
	Password string
	Db       string
	AdminDb  string // 管理后台数据库
	Schema   string // postgres 专用
}

type RedisSettingS struct {
	Host     string
	Port     string
	Password string
	Db       int
}

type RabbitmqSettingS struct {
	Host     string
	Port     string
	User     string
	Password string
	Vhost    string
}

var MysqlSettings = &MysqlSettingS{}
var RedisSettings = &RedisSettingS{}
var RabbitmqSettings = &RabbitmqSettingS{}
