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
				&model.ConversationSession{},
				&model.KnowledgeDocument{},
				&model.ImageKnowledgeDocument{},
				&model.EmbeddingTask{},
			},
		},
	}
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
	}

	return nil
}
