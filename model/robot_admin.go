package model

type RobotAdmin struct {
	RobotID    string `gorm:"column:robot_id;primaryKey;type:varchar(32);not null" json:"robot_id"`
	Owner      string `gorm:"column:owner;type:varchar(15);not null" json:"owner"`
	WxID       string `gorm:"column:wxid;type:varchar(32);not null" json:"wxid"`
	DeviceID   string `gorm:"column:device_id;type:varchar(32);not null" json:"device_id"`
	DeviceName string `gorm:"column:device_name;type:varchar(32);not null" json:"device_name"`
	ServerHost string `gorm:"column:server_host;type:varchar(32);not null" json:"server_host"`
	ServerPort int    `gorm:"column:server_port;type:int;not null" json:"server_port"`
}

func (RobotAdmin) TableName() string {
	return "robot"
}
