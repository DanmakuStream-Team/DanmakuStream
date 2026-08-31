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
		&model.VideoCollection{},
		&model.VideoCollectionItem{},
		&model.VideoCollaborator{},
		&model.Danmaku{},
		&model.Comment{},
		&model.LiveRoom{},
		&model.LiveLike{},
		&model.LiveGift{},
		&model.LiveReplay{},
		&model.DynamicPost{},
		&model.LiveSchedule{},
		&model.LiveReservation{},
		&model.Notification{},
		&model.SiteBanner{},
		&model.SiteAnnouncement{},
		&model.TrafficStat{},
		&model.CreatorDailyStat{},
		&model.VideoDailyStat{},
		&model.FollowGroup{},
		&model.Follow{},
		&model.UserBlock{},
		&model.ChatMessage{},
		&model.CreatorMembershipPlan{},
		&model.CreatorSubscription{},
		&model.SubscriptionOrder{},
		&model.Like{},
		&model.Collect{},
		&model.CommentLike{},
		&model.WatchHistory{},
		&model.WatchLater{},
	)
	// Existing rows predate the independent transcode state. A non-empty media
	// URL means the legacy record completed processing; otherwise it is pending.
	db.Model(&model.Video{}).
		Where("transcode_status IS NULL OR transcode_status = ''").
		Update("transcode_status", gorm.Expr("CASE WHEN video_url IS NOT NULL AND video_url <> '' THEN 'ready' ELSE 'processing' END"))

	videoDir := c.VideoDir
	if videoDir == "" {
		videoDir = "data"
	}
	absDir, _ := filepath.Abs(videoDir)
	os.MkdirAll(filepath.Join(absDir, "videos"), 0755)
	os.MkdirAll(filepath.Join(absDir, "covers"), 0755)
	os.MkdirAll(filepath.Join(absDir, "avatars"), 0755)
	os.MkdirAll(filepath.Join(absDir, "images"), 0755)
	os.MkdirAll(filepath.Join(absDir, "live"), 0755)

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
