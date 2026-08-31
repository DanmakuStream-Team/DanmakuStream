package database

import (
	"database/sql"
	"fmt"

	"danmakustream/engagement-service/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn), DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		return nil, fmt.Errorf("connect engagement_db: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(30)
	if err := db.AutoMigrate(model.OwnedTables()...); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migrate engagement_db: %w", err)
	}
	return db, nil
}

func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
func SQL(db *gorm.DB) (*sql.DB, error) { return db.DB() }
