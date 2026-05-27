package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"wechat-robot-client/vars"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestChatRoom(t *testing.T) {
	// 创建机器人实例连接对象
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local", "root", "", "", "3306", "")
	mysqlConfig := mysql.Config{
		DSN:                     dsn,
		DontSupportRenameIndex:  true, // 重命名索引时采用删除并新建的方式
		DontSupportRenameColumn: true, // 用 `change` 重命名列
	}
	// gorm 配置
	gormConfig := gorm.Config{}
	// 是否开启调试模式
	if flag, _ := strconv.ParseBool(os.Getenv("GORM_DEBUG")); flag {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}
	var err error
	vars.DB, err = gorm.Open(mysql.New(mysqlConfig), &gormConfig)
	if err != nil {
		t.Fatalf("连接数据库失败: %v", err)
	}
	svc := NewChatRoomService(context.Background())
	svc.ChatRoomAISummary()
}
