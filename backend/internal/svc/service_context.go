package svc

import (
	"danmakustream/backend/internal/config"
	"danmakustream/backend/internal/model/mysql"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ServiceContext holds shared dependencies injected into handlers/logic.
type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	MinIO  *minio.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := initDB(c)
	rdb := initRedis(c)
	mc := initMinIO(c)

	// Auto migrate tables
	db.AutoMigrate(
		&model.User{},
		&model.Video{},
		&model.Danmaku{},
		&model.Comment{},
		&model.LiveRoom{},
		&model.Follow{},
		&model.Like{},
		&model.Collect{},
	)

	return &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  rdb,
		MinIO:  mc,
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

func initRedis(c config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: c.Redis.Host,
	})
	return rdb
}

func initMinIO(c config.Config) *minio.Client {
	mc, err := minio.New(c.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinIO.AccessKey, c.MinIO.SecretKey, ""),
		Secure: c.MinIO.UseSSL,
	})
	if err != nil {
		panic("failed to connect to MinIO: " + err.Error())
	}
	return mc
}
