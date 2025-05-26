package shutdown

import (
	"context"

	"gorm.io/gorm"
)

type DBConnection struct {
	DB      *gorm.DB
	AdminDB *gorm.DB
}

func (d *DBConnection) Name() string {
	return "mysql 数据库连接"
}

func (d *DBConnection) Shutdown(ctx context.Context) error {
	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
	}

	if d.AdminDB != nil {
		adminSqlDB, err := d.AdminDB.DB()
		if err != nil {
			return err
		}
		if err := adminSqlDB.Close(); err != nil {
			return err
		}
	}
	return nil
}
