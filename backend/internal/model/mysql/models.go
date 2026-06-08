package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string `gorm:"uniqueIndex;size:50;not null"`
	Password    string `gorm:"not null"` // bcrypt hash
	Nickname    string `gorm:"uniqueIndex;size:50;not null"`
	Avatar      string `gorm:"size:500"`
	Bio         string `gorm:"size:500"`
	Role        string `gorm:"size:20;default:user"` // user | creator | admin
	FollowCount int64  `gorm:"default:0"`
	FanCount    int64  `gorm:"default:0"`
}

type Video struct {
	gorm.Model
	Title        string `gorm:"size:200;not null"`
	Description  string `gorm:"type:text"`
	CoverURL     string `gorm:"size:500"`
	VideoURL     string `gorm:"size:500"`
	Duration     int    `gorm:"default:0"` // seconds
	ViewCount    int64  `gorm:"default:0"`
	LikeCount    int64  `gorm:"default:0"`
	CollectCount int64  `gorm:"default:0"`
	DanmakuCount int64  `gorm:"default:0"`
	Status       string `gorm:"size:20;default:pending"` // pending | approved | rejected
	AuthorID     uint   `gorm:"not null;index"`
	Author       User   `gorm:"foreignKey:AuthorID"`
	Tags         string `gorm:"size:500"` // comma-separated
	Category     string `gorm:"column:category;type:varchar(32)" json:"category"`
}

type Danmaku struct {
	gorm.Model
	VideoID  uint   `gorm:"not null;index"`
	UserID   uint   `gorm:"not null;index"`
	Content  string `gorm:"size:200;not null"`
	Time     int    `gorm:"not null"` // seconds offset in video
	Color    string `gorm:"size:10;default:#FFFFFF"`
	FontSize string `gorm:"size:10;default:medium"` // small | medium | large
	Type     string `gorm:"size:10;default:scroll"` // scroll | top | bottom
	Blocked  bool   `gorm:"default:false"`
}

type Comment struct {
	gorm.Model
	VideoID   uint      `gorm:"not null;index"`
	UserID    uint      `gorm:"not null;index"`
	ParentID  *uint     `gorm:"index"` // nil = top-level comment
	Content   string    `gorm:"type:text;not null"`
	LikeCount int64     `gorm:"default:0"`
	User      User      `gorm:"foreignKey:UserID"`
	Replies   []Comment `gorm:"foreignKey:ParentID"`
}

type LiveRoom struct {
	gorm.Model
	Title       string `gorm:"size:200;not null"`
	CoverURL    string `gorm:"size:500"`
	StreamKey   string `gorm:"uniqueIndex;size:100"`
	Status      string `gorm:"size:20;default:idle"` // idle | live | ended
	ViewerCount int64  `gorm:"default:0"`
	OwnerID     uint   `gorm:"not null;uniqueIndex"`
	Owner       User   `gorm:"foreignKey:OwnerID"`
	StartedAt   *time.Time
	EndedAt     *time.Time
}

type DynamicPost struct {
	gorm.Model
	UserID  uint   `gorm:"not null;index"`
	User    User   `gorm:"foreignKey:UserID"`
	Content string `gorm:"type:text;not null"`
	Images  string `gorm:"size:1000"`
}

type LiveSchedule struct {
	gorm.Model
	Title         string `gorm:"size:200;not null"`
	CoverURL      string `gorm:"size:500"`
	ScheduledAt   time.Time
	Status        string `gorm:"size:20;default:pending"` // pending | canceled | live
	ReminderCount int64  `gorm:"default:0"`
	OwnerID       uint   `gorm:"not null;index"`
	Owner         User   `gorm:"foreignKey:OwnerID"`
}

type LiveReservation struct {
	gorm.Model
	ScheduleID uint         `gorm:"not null;uniqueIndex:idx_live_reservation"`
	UserID     uint         `gorm:"not null;uniqueIndex:idx_live_reservation"`
	Schedule   LiveSchedule `gorm:"foreignKey:ScheduleID"`
	User       User         `gorm:"foreignKey:UserID"`
}

type Notification struct {
	gorm.Model
	UserID  uint   `gorm:"not null;index"`
	ActorID *uint  `gorm:"index"`
	Type    string `gorm:"size:50;not null"`
	Title   string `gorm:"size:200;not null"`
	Content string `gorm:"type:text"`
	Link    string `gorm:"size:500"`
	Read    bool   `gorm:"default:false;index"`
	User    User   `gorm:"foreignKey:UserID"`
	Actor   User   `gorm:"foreignKey:ActorID"`
}

// UserInfo is a safe DTO returned to the client (no password field).
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

type Follow struct {
	gorm.Model
	FollowerID uint `gorm:"not null;uniqueIndex:idx_follow"`
	FolloweeID uint `gorm:"not null;uniqueIndex:idx_follow"`
}

type Like struct {
	gorm.Model
	UserID  uint `gorm:"not null;uniqueIndex:idx_like"`
	VideoID uint `gorm:"not null;uniqueIndex:idx_like"`
}

type Collect struct {
	gorm.Model
	UserID  uint `gorm:"not null;uniqueIndex:idx_collect"`
	VideoID uint `gorm:"not null;uniqueIndex:idx_collect"`
}

type CommentLike struct {
	gorm.Model
	UserID    uint `gorm:"not null;uniqueIndex:idx_comment_like"`
	CommentID uint `gorm:"not null;uniqueIndex:idx_comment_like"`
}
