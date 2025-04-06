package gtool

import (
	"context"

	"gorm.io/gorm"
)

func WithOrmContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if db == nil {
		return nil
	}

	return db.WithContext(ctx)
}
