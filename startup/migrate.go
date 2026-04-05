package startup

import (
	"fmt"
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"

	"gorm.io/gorm"
)

type migrateTask struct {
	name   string
	db     func() *gorm.DB
	models []any
}

func autoMigrateTasks() []migrateTask {
	return []migrateTask{
		{
			name: "robot",
			db: func() *gorm.DB {
				return vars.DB
			},
			models: []any{
				&model.GlobalSettings{},
				&model.Memory{},
				&model.UserProfile{},
				&model.ConversationSession{},
				&model.KnowledgeDocument{},
				&model.ImageKnowledgeDocument{},
				&model.EmbeddingTask{},
				&model.KnowledgeCategory{},
				&model.OSSSettings{},
			},
		},
	}
}

type enumMigration struct {
	table  string
	column string
	sql    string
}

func enumMigrations() []enumMigration {
	return []enumMigration{
		{
			table:  "oss_settings",
			column: "oss_provider",
			sql:    "ALTER TABLE oss_settings MODIFY COLUMN oss_provider ENUM('aliyun','tencent_cloud','cloudflare','volcengine') NOT NULL DEFAULT 'aliyun' COMMENT '对象存储服务商'",
		},
		{
			table:  "oss_settings",
			column: "auto_upload_image_mode",
			sql:    "ALTER TABLE oss_settings MODIFY COLUMN auto_upload_image_mode ENUM('all','ai_only') NOT NULL DEFAULT 'ai_only' COMMENT '自动上传图片模式'",
		},
		{
			table:  "oss_settings",
			column: "auto_upload_video_mode",
			sql:    "ALTER TABLE oss_settings MODIFY COLUMN auto_upload_video_mode ENUM('all','ai_only') NOT NULL DEFAULT 'ai_only' COMMENT '自动上传视频模式'",
		},
		{
			table:  "oss_settings",
			column: "auto_upload_file_mode",
			sql:    "ALTER TABLE oss_settings MODIFY COLUMN auto_upload_file_mode ENUM('all','ai_only') NOT NULL DEFAULT 'ai_only' COMMENT '自动上传文件模式'",
		},
	}
}

func migrateEnumColumns(db *gorm.DB) error {
	for _, m := range enumMigrations() {
		// 仅当表存在时才执行
		if !db.Migrator().HasTable(m.table) {
			continue
		}
		if err := db.Exec(m.sql).Error; err != nil {
			return fmt.Errorf("枚举迁移失败 [%s.%s]: %w", m.table, m.column, err)
		}
		log.Printf("[enum migrate] %s.%s 迁移完成", m.table, m.column)
	}
	return nil
}

func AutoMigrate() error {
	for _, task := range autoMigrateTasks() {
		if len(task.models) == 0 {
			continue
		}

		db := task.db()
		if db == nil {
			return fmt.Errorf("%s 数据库未初始化", task.name)
		}

		if err := db.AutoMigrate(task.models...); err != nil {
			return fmt.Errorf("%s 数据库自动迁移失败: %w", task.name, err)
		}

		log.Printf("[%s] 自动迁移完成，当前表数量: %d", task.name, len(task.models))

		if err := migrateEnumColumns(db); err != nil {
			return err
		}
	}

	return nil
}
