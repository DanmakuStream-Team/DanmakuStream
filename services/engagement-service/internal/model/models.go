package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel keeps GORM's conventional persistence fields while exposing the
// camelCase JSON contract consumed by the gateway/frontend. Embedding
// gorm.Model directly serializes ID/CreatedAt/... with Go field names, which
// made otherwise successful microservice responses unusable by the browser.
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type VideoLike struct {
	BaseModel
	UserID  uint `gorm:"not null;uniqueIndex:idx_video_like" json:"userId"`
	VideoID uint `gorm:"not null;uniqueIndex:idx_video_like" json:"videoId"`
}

func (VideoLike) TableName() string { return "video_likes" }

type VideoCollection struct {
	BaseModel
	UserID  uint `gorm:"not null;uniqueIndex:idx_video_collection" json:"userId"`
	VideoID uint `gorm:"not null;uniqueIndex:idx_video_collection" json:"videoId"`
}

func (VideoCollection) TableName() string { return "video_collections" }

type Comment struct {
	BaseModel
	VideoID   uint   `gorm:"not null;index" json:"videoId"`
	UserID    uint   `gorm:"not null;index" json:"userId"`
	ParentID  *uint  `gorm:"index" json:"parentId,omitempty"`
	Content   string `gorm:"type:text;not null" json:"content"`
	LikeCount int64  `gorm:"not null;default:0" json:"likeCount"`
}
type CommentLike struct {
	BaseModel
	UserID    uint `gorm:"not null;uniqueIndex:idx_comment_like" json:"userId"`
	CommentID uint `gorm:"not null;uniqueIndex:idx_comment_like" json:"commentId"`
}
type Danmaku struct {
	BaseModel
	VideoID  uint   `gorm:"not null;index" json:"videoId"`
	Scene    string `gorm:"size:20;not null;default:video;index" json:"scene"`
	UserID   uint   `gorm:"not null;index" json:"userId"`
	Content  string `gorm:"size:500;not null" json:"content"`
	Time     int    `gorm:"not null;default:0" json:"time"`
	Color    string `gorm:"size:10;default:#FFFFFF" json:"color"`
	FontSize string `gorm:"size:10;default:medium" json:"fontSize"`
	Type     string `gorm:"size:10;default:scroll" json:"type"`
	Blocked  bool   `gorm:"default:false;index" json:"blocked"`
}

func (Danmaku) TableName() string { return "danmaku" }

type WatchHistory struct {
	BaseModel
	UserID   uint `gorm:"not null;uniqueIndex:idx_watch_history" json:"userId"`
	VideoID  uint `gorm:"not null;uniqueIndex:idx_watch_history" json:"videoId"`
	Position int  `gorm:"not null;default:0" json:"position"`
}
type WatchLater struct {
	BaseModel
	UserID  uint `gorm:"not null;uniqueIndex:idx_watch_later" json:"userId"`
	VideoID uint `gorm:"not null;uniqueIndex:idx_watch_later" json:"videoId"`
}
type LiveRoom struct {
	BaseModel
	Title           string     `gorm:"size:200;not null" json:"title"`
	CoverURL        string     `gorm:"size:500" json:"coverUrl"`
	StreamKey       string     `gorm:"size:100;not null;uniqueIndex" json:"streamKey,omitempty"`
	Status          string     `gorm:"size:20;not null;default:idle;index" json:"status"`
	OwnerID         uint       `gorm:"not null;uniqueIndex" json:"ownerId"`
	ViewerCount     int64      `gorm:"not null;default:0" json:"viewerCount"`
	ViewerPeak      int64      `gorm:"not null;default:0" json:"viewerPeak"`
	LikeCount       int64      `gorm:"not null;default:0" json:"likeCount"`
	GiftValue       int64      `gorm:"not null;default:0" json:"giftValue"`
	ChatMode        string     `gorm:"size:20;not null;default:everyone" json:"chatMode"`
	SlowModeSeconds int        `gorm:"not null;default:0" json:"slowModeSeconds"`
	PinnedMessage   string     `gorm:"size:500" json:"pinnedMessage"`
	StartedAt       *time.Time `json:"startedAt,omitempty"`
	EndedAt         *time.Time `json:"endedAt,omitempty"`
}
type LiveSchedule struct {
	BaseModel
	Title         string    `gorm:"size:200;not null" json:"title"`
	CoverURL      string    `gorm:"size:500" json:"coverUrl"`
	ScheduledAt   time.Time `gorm:"not null;index" json:"scheduledAt"`
	Status        string    `gorm:"size:20;not null;default:pending;index" json:"status"`
	ReminderCount int64     `gorm:"not null;default:0" json:"reminderCount"`
	OwnerID       uint      `gorm:"not null;index" json:"ownerId"`
}
type LiveReservation struct {
	BaseModel
	ScheduleID uint `gorm:"not null;uniqueIndex:idx_live_reservation" json:"scheduleId"`
	UserID     uint `gorm:"not null;uniqueIndex:idx_live_reservation" json:"userId"`
}
type LiveGift struct {
	BaseModel
	RoomID  uint   `gorm:"not null;index" json:"roomId"`
	UserID  uint   `gorm:"not null;index" json:"userId"`
	GiftKey string `gorm:"size:30;not null" json:"giftKey"`
	Name    string `gorm:"size:40;not null" json:"name"`
	Count   int    `gorm:"not null" json:"count"`
	Value   int64  `gorm:"not null" json:"value"`
	Message string `gorm:"size:500" json:"message"`
}
type LiveLike struct {
	BaseModel
	RoomID uint `gorm:"not null;uniqueIndex:idx_live_like" json:"roomId"`
	UserID uint `gorm:"not null;uniqueIndex:idx_live_like" json:"userId"`
}
type SuperChat struct {
	BaseModel
	RoomID         uint   `gorm:"not null;index" json:"roomId"`
	UserID         uint   `gorm:"not null;index" json:"userId"`
	GiftID         uint   `gorm:"not null;uniqueIndex" json:"giftId"`
	Content        string `gorm:"size:500;not null" json:"content"`
	DisplaySeconds int    `gorm:"not null" json:"displaySeconds"`
}
type LiveInteraction struct {
	BaseModel
	RoomID uint   `gorm:"not null;uniqueIndex:idx_live_interaction" json:"roomId"`
	UserID uint   `gorm:"not null;uniqueIndex:idx_live_interaction" json:"userId"`
	Kind   string `gorm:"size:20;not null;uniqueIndex:idx_live_interaction" json:"kind"`
	Value  int64  `gorm:"not null;default:0" json:"value"`
}

func OwnedTables() []any {
	return []any{&VideoLike{}, &VideoCollection{}, &Comment{}, &CommentLike{}, &Danmaku{}, &WatchHistory{}, &WatchLater{}, &LiveRoom{}, &LiveSchedule{}, &LiveReservation{}, &LiveGift{}, &LiveLike{}, &SuperChat{}, &LiveInteraction{}}
}
