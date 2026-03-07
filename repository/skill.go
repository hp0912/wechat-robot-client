package repository

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/skills"
)

type SkillRepo struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewSkillRepo(ctx context.Context, db *gorm.DB) *SkillRepo {
	return &SkillRepo{Ctx: ctx, DB: db}
}

func (r *SkillRepo) AutoMigrate() error {
	return r.DB.WithContext(r.Ctx).AutoMigrate(&model.Skill{})
}

func (r *SkillRepo) Create(s *model.Skill) error {
	return r.DB.WithContext(r.Ctx).Create(s).Error
}

func (r *SkillRepo) Update(s *model.Skill) error {
	return r.DB.WithContext(r.Ctx).Save(s).Error
}

func (r *SkillRepo) Delete(name string) error {
	return r.DB.WithContext(r.Ctx).Where("name = ?", name).Delete(&model.Skill{}).Error
}

func (r *SkillRepo) FindByName(name string) (*model.Skill, error) {
	var s model.Skill
	err := r.DB.WithContext(r.Ctx).Where("name = ?", name).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SkillRepo) FindAll() ([]*model.Skill, error) {
	var list []*model.Skill
	err := r.DB.WithContext(r.Ctx).Find(&list).Error
	return list, err
}

func (r *SkillRepo) Upsert(s *model.Skill) error {
	existing, err := r.FindByName(s.Name)
	if err != nil {
		return err
	}
	if existing != nil {
		s.ID = existing.ID
		return r.Update(s)
	}
	return r.Create(s)
}

// ToSkillSource 从 model.Skill 解析出 skills.SkillSource
func ToSkillSource(s *model.Skill) skills.SkillSource {
	var source skills.SkillSource
	if s.Source != nil {
		_ = json.Unmarshal([]byte(s.Source), &source)
	}
	return source
}

// SourceToJSON 将 skills.SkillSource 序列化为 datatypes.JSON
func SourceToJSON(src skills.SkillSource) []byte {
	data, _ := json.Marshal(src)
	return data
}
