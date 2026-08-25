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
	Role        string `gorm:"size:20;default:user"` // user | creator | moderator | admin
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

type VideoCollection struct {
	gorm.Model
	Title       string `gorm:"size:120;not null"`
	Description string `gorm:"size:500"`
	CoverURL    string `gorm:"size:500"`
	OwnerID     uint   `gorm:"not null;index"`
	Owner       User   `gorm:"foreignKey:OwnerID"`
}

type VideoCollectionItem struct {
	gorm.Model
	CollectionID uint            `gorm:"not null;uniqueIndex:idx_collection_video"`
	VideoID      uint            `gorm:"not null;uniqueIndex:idx_collection_video"`
	Sort         int             `gorm:"default:0"`
	Collection   VideoCollection `gorm:"foreignKey:CollectionID"`
	Video        Video           `gorm:"foreignKey:VideoID"`
}

type VideoCollaborator struct {
	gorm.Model
	VideoID uint  `gorm:"not null;uniqueIndex:idx_video_collaborator"`
	UserID  uint  `gorm:"not null;uniqueIndex:idx_video_collaborator"`
	Video   Video `gorm:"foreignKey:VideoID"`
	User    User  `gorm:"foreignKey:UserID"`
}

type Danmaku struct {
	gorm.Model
	VideoID  uint   `gorm:"not null;index"`
	Scene    string `gorm:"size:20;default:video;index"` // video | live
	UserID   uint   `gorm:"not null;index"`
	Content  string `gorm:"size:500;not null"`
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
	Title           string `gorm:"size:200;not null"`
	CoverURL        string `gorm:"size:500"`
	StreamKey       string `gorm:"uniqueIndex;size:100"`
	Status          string `gorm:"size:20;default:idle"` // idle | live | ended
	ViewerCount     int64  `gorm:"default:0"`
	ViewerPeak      int64  `gorm:"default:0"`
	LikeCount       int64  `gorm:"default:0"`
	GiftValue       int64  `gorm:"default:0"`
	ChatMode        string `gorm:"size:20;not null;default:everyone"` // everyone | followers | members
	SlowModeSeconds int    `gorm:"not null;default:0"`
	PinnedMessage   string `gorm:"size:500"`
	OwnerID         uint   `gorm:"not null;uniqueIndex"`
	Owner           User   `gorm:"foreignKey:OwnerID"`
	StartedAt       *time.Time
	EndedAt         *time.Time
}

// LiveLike records one viewer's like state in a live room.
type LiveLike struct {
	gorm.Model
	RoomID uint `gorm:"not null;uniqueIndex:idx_live_like"`
	UserID uint `gorm:"not null;uniqueIndex:idx_live_like"`
}

// LiveGift keeps an auditable gift event. Value is a virtual support score;
// real-money settlement can be layered on top without changing room ranking.
type LiveGift struct {
	gorm.Model
	RoomID         uint   `gorm:"not null;index"`
	UserID         uint   `gorm:"not null;index"`
	GiftKey        string `gorm:"size:30;not null;index"`
	Name           string `gorm:"size:40;not null"`
	Count          int    `gorm:"not null"`
	Value          int64  `gorm:"not null"`
	Message        string `gorm:"size:500"`
	DisplaySeconds int    `gorm:"not null;default:0"`
	User           User   `gorm:"foreignKey:UserID"`
}

// LiveReplay is an immutable archive for one completed broadcast. LiveRoom is
// reused by its owner, so replay metadata must be stored separately.
type LiveReplay struct {
	gorm.Model
	RoomID     uint      `gorm:"not null;index"`
	Title      string    `gorm:"size:200;not null"`
	CoverURL   string    `gorm:"size:500"`
	StreamKey  string    `gorm:"size:100;not null;uniqueIndex"`
	ReplayURL  string    `gorm:"size:500"`
	Status     string    `gorm:"size:20;not null;default:processing;index"` // processing | ready | unavailable
	Duration   int       `gorm:"default:0"`
	ViewerPeak int64     `gorm:"default:0"`
	OwnerID    uint      `gorm:"not null;index"`
	Owner      User      `gorm:"foreignKey:OwnerID"`
	StartedAt  time.Time `gorm:"not null;index"`
	EndedAt    time.Time `gorm:"not null"`
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

type SiteBanner struct {
	gorm.Model
	Title    string `gorm:"size:120;not null"`
	ImageURL string `gorm:"size:500"`
	Link     string `gorm:"size:500"`
	Enabled  bool   `gorm:"default:true;index"`
	Sort     int    `gorm:"default:0"`
}

type SiteAnnouncement struct {
	gorm.Model
	Content   string `gorm:"size:500;not null"`
	Enabled   bool   `gorm:"default:true;index"`
	StartedAt *time.Time
	EndedAt   *time.Time
}

type TrafficStat struct {
	gorm.Model
	Date     string `gorm:"size:10;not null;uniqueIndex:idx_traffic_date_category"` // YYYY-MM-DD
	Category string `gorm:"size:32;not null;uniqueIndex:idx_traffic_date_category"`
	Bytes    uint64 `gorm:"default:0"`
}

// CreatorDailyStat stores creator metrics that need a historical trend.
type CreatorDailyStat struct {
	gorm.Model
	CreatorID    uint   `gorm:"not null;uniqueIndex:idx_creator_daily_stat"`
	Date         string `gorm:"size:10;not null;uniqueIndex:idx_creator_daily_stat"` // YYYY-MM-DD
	ViewDelta    int64  `gorm:"default:0"`
	CollectDelta int64  `gorm:"default:0"`
	StreamCount  int64  `gorm:"default:0"`
}

// VideoDailyStat keeps per-video metrics for creator analytics.
type VideoDailyStat struct {
	gorm.Model
	CreatorID    uint   `gorm:"not null;index"`
	VideoID      uint   `gorm:"not null;uniqueIndex:idx_video_daily_stat"`
	Date         string `gorm:"size:10;not null;uniqueIndex:idx_video_daily_stat"` // YYYY-MM-DD
	ViewDelta    int64  `gorm:"default:0"`
	CollectDelta int64  `gorm:"default:0"`
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
	FollowerID uint  `gorm:"not null;uniqueIndex:idx_follow"`
	FolloweeID uint  `gorm:"not null;uniqueIndex:idx_follow"`
	GroupID    *uint `gorm:"index"`
	Special    bool  `gorm:"default:false;index"`
}

type FollowGroup struct {
	gorm.Model
	OwnerID uint   `gorm:"not null;index"`
	Name    string `gorm:"size:30;not null"`
}

type UserBlock struct {
	gorm.Model
	BlockerID uint `gorm:"not null;uniqueIndex:idx_user_block"`
	BlockedID uint `gorm:"not null;uniqueIndex:idx_user_block"`
}

type ChatMessage struct {
	gorm.Model
	SenderID      uint   `gorm:"not null;index;index:idx_chat_pair"`
	ReceiverID    uint   `gorm:"not null;index;index:idx_chat_pair;index:idx_chat_receiver_read"`
	MessageType   string `gorm:"size:20;not null;default:text;index"` // text | image | video | video_share
	Content       string `gorm:"type:text;not null"`
	MediaURL      string `gorm:"size:500"`
	MediaName     string `gorm:"size:255"`
	SharedVideoID *uint  `gorm:"index"`
	Read          bool   `gorm:"default:false;index;index:idx_chat_receiver_read"`
	Sender        User   `gorm:"foreignKey:SenderID"`
	Receiver      User   `gorm:"foreignKey:ReceiverID"`
	SharedVideo   Video  `gorm:"foreignKey:SharedVideoID"`
}

type CreatorMembershipPlan struct {
	gorm.Model
	CreatorID  uint   `gorm:"not null;uniqueIndex"`
	PriceCents int64  `gorm:"not null;default:500"`
	Benefits   string `gorm:"size:500"`
	Enabled    bool   `gorm:"default:false;index"`
	Creator    User   `gorm:"foreignKey:CreatorID"`
}

type CreatorSubscription struct {
	gorm.Model
	SubscriberID uint      `gorm:"not null;uniqueIndex:idx_creator_subscription"`
	CreatorID    uint      `gorm:"not null;uniqueIndex:idx_creator_subscription"`
	PriceCents   int64     `gorm:"not null"`
	Status       string    `gorm:"size:20;not null;default:active;index"` // active | canceled | expired
	AutoRenew    bool      `gorm:"default:false"`
	StartedAt    time.Time `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null;index"`
	Subscriber   User      `gorm:"foreignKey:SubscriberID"`
	Creator      User      `gorm:"foreignKey:CreatorID"`
}

type SubscriptionOrder struct {
	gorm.Model
	OrderNo      string     `gorm:"size:40;not null;uniqueIndex"`
	SubscriberID uint       `gorm:"not null;index"`
	CreatorID    uint       `gorm:"not null;index"`
	AmountCents  int64      `gorm:"not null"`
	Months       int        `gorm:"not null"`
	Status       string     `gorm:"size:20;not null;default:pending;index"` // pending | paid | canceled
	PaidAt       *time.Time `gorm:"index"`
	Subscriber   User       `gorm:"foreignKey:SubscriberID"`
	Creator      User       `gorm:"foreignKey:CreatorID"`
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

type WatchHistory struct {
	gorm.Model
	UserID   uint  `gorm:"not null;uniqueIndex:idx_watch_history"`
	VideoID  uint  `gorm:"not null;uniqueIndex:idx_watch_history"`
	Position int   `gorm:"default:0"` // seconds
	Video    Video `gorm:"foreignKey:VideoID"`
}

type WatchLater struct {
	gorm.Model
	UserID  uint  `gorm:"not null;uniqueIndex:idx_watch_later"`
	VideoID uint  `gorm:"not null;uniqueIndex:idx_watch_later"`
	Video   Video `gorm:"foreignKey:VideoID"`
}
