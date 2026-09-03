package svc

import (
	"danmakustream/user-service/internal/client"
	"danmakustream/user-service/internal/config"
	model "danmakustream/user-service/internal/model/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

type ServiceContext struct {
	Config   config.Config
	DB       *gorm.DB
	VideoDir string
	Content  *client.ContentClient
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	db, err := gorm.Open(mysql.Open(c.Database.DataSource), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.User{}, &model.Follow{}, &model.FollowGroup{}, &model.UserBlock{}, &model.CreatorMembershipPlan{}, &model.CreatorSubscription{}, &model.SubscriptionOrder{}, &model.ChatMessage{}, &model.Notification{}); err != nil {
		return nil, err
	}
	dir := c.VideoDir
	if dir == "" {
		dir = "data"
	}
	dir, _ = filepath.Abs(dir)
	_ = os.MkdirAll(filepath.Join(dir, "avatars"), 0755)
	_ = os.MkdirAll(filepath.Join(dir, "messages"), 0755)
	return &ServiceContext{Config: c, DB: db, VideoDir: dir, Content: client.NewContent(c.ContentServiceURL, c.InternalToken, c.RequestTimeout)}, nil
}
