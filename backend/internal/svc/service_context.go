package svc

import (
	"os"
	"path/filepath"

	"danmakustream/backend/internal/config"
	model "danmakustream/backend/internal/model/mysql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServiceContext struct {
	Config   config.Config
	DB       *gorm.DB
	VideoDir string
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := initDB(c)

	db.AutoMigrate(
		&model.User{},
		&model.Video{},
		&model.Danmaku{},
		&model.Comment{},
		&model.LiveRoom{},
		&model.DynamicPost{},
		&model.LiveSchedule{},
		&model.LiveReservation{},
		&model.Notification{},
		&model.Follow{},
		&model.Like{},
		&model.Collect{},
		&model.CommentLike{},
	)

	videoDir := c.VideoDir
	if videoDir == "" {
		videoDir = "data"
	}
	absDir, _ := filepath.Abs(videoDir)
	os.MkdirAll(filepath.Join(absDir, "videos"), 0755)
	os.MkdirAll(filepath.Join(absDir, "covers"), 0755)
	os.MkdirAll(filepath.Join(absDir, "avatars"), 0755)
	os.MkdirAll(filepath.Join(absDir, "images"), 0755)

	return &ServiceContext{
		Config:   c,
		DB:       db,
		VideoDir: absDir,
	}
}

func initDB(c config.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.Database.DataSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	return db
}
