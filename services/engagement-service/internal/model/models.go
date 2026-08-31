package model

import (
	"gorm.io/gorm"
	"time"
)

type VideoLike struct {
	gorm.Model
	UserID  uint `gorm:"not null;uniqueIndex:idx_video_like"`
	VideoID uint `gorm:"not null;uniqueIndex:idx_video_like"`
}

func (VideoLike) TableName() string { return "video_likes" }

type VideoCollection struct {
	gorm.Model
	UserID  uint `gorm:"not null;uniqueIndex:idx_video_collection"`
	VideoID uint `gorm:"not null;uniqueIndex:idx_video_collection"`
}

func (VideoCollection) TableName() string { return "video_collections" }

type Comment struct {
	gorm.Model
	VideoID   uint   `gorm:"not null;index"`
	UserID    uint   `gorm:"not null;index"`
	ParentID  *uint  `gorm:"index"`
	Content   string `gorm:"type:text;not null"`
	LikeCount int64  `gorm:"not null;default:0"`
}
type CommentLike struct {
	gorm.Model
	UserID    uint `gorm:"not null;uniqueIndex:idx_comment_like"`
	CommentID uint `gorm:"not null;uniqueIndex:idx_comment_like"`
}
type Danmaku struct {
	gorm.Model
	VideoID  uint   `gorm:"not null;index"`
	Scene    string `gorm:"size:20;not null;default:video;index"`
	UserID   uint   `gorm:"not null;index"`
	Content  string `gorm:"size:500;not null"`
	Time     int    `gorm:"not null;default:0"`
	Color    string `gorm:"size:10;default:#FFFFFF"`
	FontSize string `gorm:"size:10;default:medium"`
	Type     string `gorm:"size:10;default:scroll"`
	Blocked  bool   `gorm:"default:false;index"`
}

func (Danmaku) TableName() string { return "danmaku" }

type WatchHistory struct {
	gorm.Model
	UserID   uint `gorm:"not null;uniqueIndex:idx_watch_history"`
	VideoID  uint `gorm:"not null;uniqueIndex:idx_watch_history"`
	Position int  `gorm:"not null;default:0"`
}
type WatchLater struct {
	gorm.Model
	UserID  uint `gorm:"not null;uniqueIndex:idx_watch_later"`
	VideoID uint `gorm:"not null;uniqueIndex:idx_watch_later"`
}
type LiveRoom struct {
	gorm.Model
	Title           string `gorm:"size:200;not null"`
	CoverURL        string `gorm:"size:500"`
	StreamKey       string `gorm:"size:100;not null;uniqueIndex"`
	Status          string `gorm:"size:20;not null;default:idle;index"`
	OwnerID         uint   `gorm:"not null;uniqueIndex"`
	ViewerCount     int64  `gorm:"not null;default:0"`
	ViewerPeak      int64  `gorm:"not null;default:0"`
	LikeCount       int64  `gorm:"not null;default:0"`
	GiftValue       int64  `gorm:"not null;default:0"`
	ChatMode        string `gorm:"size:20;not null;default:everyone"`
	SlowModeSeconds int    `gorm:"not null;default:0"`
	PinnedMessage   string `gorm:"size:500"`
	StartedAt       *time.Time
	EndedAt         *time.Time
}
type LiveSchedule struct {
	gorm.Model
	Title         string    `gorm:"size:200;not null"`
	CoverURL      string    `gorm:"size:500"`
	ScheduledAt   time.Time `gorm:"not null;index"`
	Status        string    `gorm:"size:20;not null;default:pending;index"`
	ReminderCount int64     `gorm:"not null;default:0"`
	OwnerID       uint      `gorm:"not null;index"`
}
type LiveReservation struct {
	gorm.Model
	ScheduleID uint `gorm:"not null;uniqueIndex:idx_live_reservation"`
	UserID     uint `gorm:"not null;uniqueIndex:idx_live_reservation"`
}
type LiveGift struct {
	gorm.Model
	RoomID  uint   `gorm:"not null;index"`
	UserID  uint   `gorm:"not null;index"`
	GiftKey string `gorm:"size:30;not null"`
	Name    string `gorm:"size:40;not null"`
	Count   int    `gorm:"not null"`
	Value   int64  `gorm:"not null"`
	Message string `gorm:"size:500"`
}
type LiveLike struct {
	gorm.Model
	RoomID uint `gorm:"not null;uniqueIndex:idx_live_like"`
	UserID uint `gorm:"not null;uniqueIndex:idx_live_like"`
}
type SuperChat struct {
	gorm.Model
	RoomID         uint   `gorm:"not null;index"`
	UserID         uint   `gorm:"not null;index"`
	GiftID         uint   `gorm:"not null;uniqueIndex"`
	Content        string `gorm:"size:500;not null"`
	DisplaySeconds int    `gorm:"not null"`
}
type LiveInteraction struct {
	gorm.Model
	RoomID uint   `gorm:"not null;uniqueIndex:idx_live_interaction"`
	UserID uint   `gorm:"not null;uniqueIndex:idx_live_interaction"`
	Kind   string `gorm:"size:20;not null;uniqueIndex:idx_live_interaction"`
	Value  int64  `gorm:"not null;default:0"`
}

func OwnedTables() []any {
	return []any{&VideoLike{}, &VideoCollection{}, &Comment{}, &CommentLike{}, &Danmaku{}, &WatchHistory{}, &WatchLater{}, &LiveRoom{}, &LiveSchedule{}, &LiveReservation{}, &LiveGift{}, &LiveLike{}, &SuperChat{}, &LiveInteraction{}}
}
