package gormx

import (
	"gorm.io/gorm"
)

type GormUnitOfWorkIface interface {
	BeginTran(*gorm.DB) (*gorm.DB, error)
	Rollback(*gorm.DB) error
	Commit(*gorm.DB) error
}

type GormUnitOfWork struct{}

func (g *GormUnitOfWork) BeginTran(db *gorm.DB) (*gorm.DB, error) {
	newDb := db.Begin()
	if newDb.Error != nil {
		return nil, newDb.Error
	}
	return newDb, nil
}

func (g *GormUnitOfWork) Rollback(db *gorm.DB) error {
	return db.Rollback().Error
}

func (g *GormUnitOfWork) Commit(db *gorm.DB) error {
	return db.Commit().Error
}
