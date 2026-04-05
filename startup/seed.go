package startup

import (
	"context"
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

// seedItem 描述一条种子数据
type seedItem struct {
	Code        string
	Type        model.KnowledgeCategoryType
	Name        string
	Description string
}

// builtinKnowledgeCategories 声明式定义所有系统内置的知识库分类
var builtinKnowledgeCategories = []seedItem{}

// SeedData 在迁移之后执行，幂等地插入系统内置种子数据。
// 使用 FirstOrCreate 按 code 去重，已存在的记录不会被覆盖，保证可重复调用。
func SeedData() error {
	ctx := context.Background()
	repo := repository.NewKnowledgeCategoryRepo(ctx, vars.DB)

	for _, item := range builtinKnowledgeCategories {
		category := &model.KnowledgeCategory{
			Code:        item.Code,
			Type:        item.Type,
			Name:        item.Name,
			Description: item.Description,
			IsBuiltin:   true,
		}
		if err := repo.FirstOrCreate(category); err != nil {
			return err
		}
	}

	log.Printf("[Seed] 知识库分类种子数据初始化完成，内置分类数: %d", len(builtinKnowledgeCategories))
	return nil
}
