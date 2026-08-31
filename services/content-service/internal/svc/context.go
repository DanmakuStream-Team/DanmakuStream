package svc

import (
	"fmt"

	"danmakustream/content-service/internal/config"
	"danmakustream/content-service/internal/logic"
	"danmakustream/content-service/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Context struct {
	Config config.Config
	DB     *gorm.DB
	Logic  *logic.Service
}

func New(cfg config.Config) (*Context, error) {
	dialector := mysql.New(mysql.Config{DSN: cfg.DatabaseDSN, SkipInitializeWithVersion: true})
	db, err := gorm.Open(dialector, &gorm.Config{TranslateError: true, DisableAutomaticPing: true})
	if err != nil {
		return nil, fmt.Errorf("open content database: %w", err)
	}
	if cfg.AutoMigrate {
		if err := db.AutoMigrate(model.ContentModels()...); err != nil {
			return nil, fmt.Errorf("migrate content database: %w", err)
		}
	}
	return &Context{Config: cfg, DB: db, Logic: &logic.Service{DB: db}}, nil
}

func (s *Context) Close() error {
	db, err := s.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
