package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/qdrantx"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"gorm.io/gorm"
)

var codeRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

type KnowledgeCategoryService struct {
	DB *gorm.DB
}

func NewKnowledgeCategoryService(db *gorm.DB) *KnowledgeCategoryService {
	return &KnowledgeCategoryService{DB: db}
}

func (s *KnowledgeCategoryService) Create(ctx context.Context, code string, categoryType string, name, description string) (*model.KnowledgeCategory, error) {
	if code == "" || name == "" {
		return nil, errors.New("code 和 name 不能为空")
	}
	if !codeRegexp.MatchString(code) {
		return nil, errors.New("code 必须以字母开头，并且只能包含字母、数字和下划线")
	}
	typedCategory, err := s.parseCategoryType(categoryType, false)
	if err != nil {
		return nil, err
	}

	repo := repository.NewKnowledgeCategoryRepo(ctx, s.DB)

	existing, err := repo.GetByCodeAndType(code, typedCategory)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("同类型下 code 已存在")
	}

	category := &model.KnowledgeCategory{
		Code:        code,
		Type:        typedCategory,
		Name:        name,
		Description: description,
	}
	if err := repo.Create(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *KnowledgeCategoryService) Update(ctx context.Context, id int64, name, description string) error {
	if name == "" {
		return errors.New("name 不能为空")
	}

	repo := repository.NewKnowledgeCategoryRepo(ctx, s.DB)

	existing, err := repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("分类不存在")
	}

	return repo.Update(id, name, description)
}

func (s *KnowledgeCategoryService) Delete(ctx context.Context, id int64) error {
	repo := repository.NewKnowledgeCategoryRepo(ctx, s.DB)

	existing, err := repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("分类不存在")
	}
	if existing.IsBuiltin {
		return errors.New("系统内置分类不允许删除")
	}

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		categoryRepo := repository.NewKnowledgeCategoryRepo(ctx, tx)
		knowledgeRepo := repository.NewKnowledgeDocumentRepo(ctx, tx)
		imageKnowledgeRepo := repository.NewImageKnowledgeDocumentRepo(ctx, tx)

		if existing.Type == model.KnowledgeCategoryTypeText {
			knowledgeVectorIDs, err := knowledgeRepo.GetAllVectorIDsByCategory(existing.Code)
			if err != nil {
				return fmt.Errorf("查询文本知识库向量失败: %w", err)
			}
			if err := s.deleteVectors(ctx, qdrantx.CollectionKnowledge, knowledgeVectorIDs); err != nil {
				return err
			}
			if err := knowledgeRepo.DeleteByCategory(existing.Code); err != nil {
				return fmt.Errorf("删除文本知识库失败: %w", err)
			}
		}

		if existing.Type == model.KnowledgeCategoryTypeImage {
			imageVectorIDs, err := imageKnowledgeRepo.GetAllVectorIDsByCategory(existing.Code)
			if err != nil {
				return fmt.Errorf("查询图片知识库向量失败: %w", err)
			}
			if err := s.deleteVectors(ctx, qdrantx.CollectionImageKnowledge, imageVectorIDs); err != nil {
				return err
			}
			if err := imageKnowledgeRepo.DeleteByCategory(existing.Code); err != nil {
				return fmt.Errorf("删除图片知识库失败: %w", err)
			}
		}

		return categoryRepo.Delete(id)
	})
}

func (s *KnowledgeCategoryService) deleteVectors(ctx context.Context, collection string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	if vars.QdrantClient == nil {
		log.Printf("[KnowledgeCategory] Qdrant 未初始化，跳过删除集合 %s 的 %d 条向量", collection, len(ids))
		return nil
	}
	if err := vars.QdrantClient.DeleteByIDs(ctx, collection, ids); err != nil {
		return fmt.Errorf("删除向量集合 %s 失败: %w", collection, err)
	}
	return nil
}

func (s *KnowledgeCategoryService) GetByID(ctx context.Context, id int64) (*model.KnowledgeCategory, error) {
	return repository.NewKnowledgeCategoryRepo(ctx, s.DB).GetByID(id)
}

func (s *KnowledgeCategoryService) GetByCode(ctx context.Context, code string) (*model.KnowledgeCategory, error) {
	return repository.NewKnowledgeCategoryRepo(ctx, s.DB).GetByCode(code)
}

func (s *KnowledgeCategoryService) GetByCodeAndType(ctx context.Context, code string, categoryType model.KnowledgeCategoryType) (*model.KnowledgeCategory, error) {
	return repository.NewKnowledgeCategoryRepo(ctx, s.DB).GetByCodeAndType(code, categoryType)
}

func (s *KnowledgeCategoryService) List(ctx context.Context, categoryType string) ([]*model.KnowledgeCategory, error) {
	typedCategory, err := s.parseCategoryType(categoryType, true)
	if err != nil {
		return nil, err
	}
	return repository.NewKnowledgeCategoryRepo(ctx, s.DB).List(typedCategory)
}

func (s *KnowledgeCategoryService) parseCategoryType(categoryType string, allowEmpty bool) (model.KnowledgeCategoryType, error) {
	if categoryType == "" {
		if allowEmpty {
			return "", nil
		}
		return "", errors.New("type 不能为空")
	}
	typedCategory := model.KnowledgeCategoryType(categoryType)
	if !model.IsValidKnowledgeCategoryType(typedCategory) {
		return "", errors.New("type 仅支持 text 或 image")
	}
	return typedCategory, nil
}
