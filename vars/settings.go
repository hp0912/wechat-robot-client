package vars

type MysqlSettingS struct {
	Driver   string // 使用的数据库驱动，支持 mysql、postgres
	Host     string
	Port     string
	User     string
	Password string
	Db       string
	Schema   string // postgres 专用
}

type RabbitmqSettingS struct {
	Host     string
	Port     string
	User     string
	Password string
	Vhost    string
}

var MysqlSettings = &MysqlSettingS{}
var RabbitmqSettings = &RabbitmqSettingS{}
