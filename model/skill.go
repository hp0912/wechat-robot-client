package model

import (
	"time"

	"gorm.io/datatypes"
)

// SkillSourceType Skill 来源类型
type SkillSourceType string

const (
	SkillSourceLocal SkillSourceType = "local"
	SkillSourceGit   SkillSourceType = "git"
)

// Skill GORM 模型 - 已安装的 Agent Skill 配置
type Skill struct {
	ID          uint64          `gorm:"column:id;primaryKey;autoIncrement;comment:Skill主键ID" json:"id"`
	Name        string          `gorm:"column:name;type:varchar(128);not null;uniqueIndex;comment:Skill名称" json:"name"`
	Path        string          `gorm:"column:path;type:varchar(512);not null;comment:Skill在磁盘上的绝对路径" json:"path"`
	Enabled     *bool           `gorm:"column:enabled;default:true;comment:是否启用" json:"enabled"`
	SourceType  SkillSourceType `gorm:"column:source_type;type:varchar(20);default:'local';comment:来源类型：local/git" json:"source_type"`
	Source      datatypes.JSON  `gorm:"column:source;type:json;comment:来源详情(JSON)" json:"source"`
	EnvVars     datatypes.JSON  `gorm:"column:env_vars;type:json;comment:环境变量列表(JSON)" json:"env_vars"`
	InstalledAt *time.Time      `gorm:"column:installed_at;type:datetime;comment:安装时间" json:"installed_at"`
	CreatedAt   *time.Time      `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt   *time.Time      `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

// TableName 设置表名
func (Skill) TableName() string {
	return "skills"
}

// IsEnabled 判断是否启用
func (s *Skill) IsEnabled() bool {
	return s.Enabled != nil && *s.Enabled
}
