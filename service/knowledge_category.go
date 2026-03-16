package service

import (
	"context"
	"errors"
	"regexp"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"

	"gorm.io/gorm"
)

var codeRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

type KnowledgeCategoryService struct {
	DB *gorm.DB
}

func NewKnowledgeCategoryService(db *gorm.DB) *KnowledgeCategoryService {
	return &KnowledgeCategoryService{DB: db}
}

func (s *KnowledgeCategoryService) Create(ctx context.Context, code, name, description string) (*model.KnowledgeCategory, error) {
	if code == "" || name == "" {
		return nil, errors.New("code 和 name 不能为空")
	}
	if !codeRegexp.MatchString(code) {
		return nil, errors.New("code 必须以字母开头，并且只能包含字母、数字和下划线")
	}

	repo := repository.NewKnowledgeCategoryRepo(ctx, s.DB)

	existing, err := repo.GetByCode(code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("code 已存在")
	}

	category := &model.KnowledgeCategory{
		Code:        code,
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

	return repo.Delete(id)
}

func (s *KnowledgeCategoryService) GetByID(ctx context.Context, id int64) (*model.KnowledgeCategory, error) {
	return repository.NewKnowledgeCategoryRepo(ctx, s.DB).GetByID(id)
}

func (s *KnowledgeCategoryService) GetByCode(ctx context.Context, code string) (*model.KnowledgeCategory, error) {
	return repository.NewKnowledgeCategoryRepo(ctx, s.DB).GetByCode(code)
}

func (s *KnowledgeCategoryService) List(ctx context.Context) ([]*model.KnowledgeCategory, error) {
	return repository.NewKnowledgeCategoryRepo(ctx, s.DB).List()
}
