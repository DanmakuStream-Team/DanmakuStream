package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username    string `gorm:"uniqueIndex;size:50;not null"`
	Password    string `gorm:"not null"`
	Nickname    string `gorm:"uniqueIndex;size:50;not null"`
	Avatar      string `gorm:"size:500"`
	Bio         string `gorm:"size:500"`
	Role        string `gorm:"size:20;default:user"`
	FollowCount int64  `gorm:"default:0"`
	FanCount    int64  `gorm:"default:0"`
}
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
type ChatMessage struct {
	gorm.Model
	SenderID        uint    `gorm:"not null;index;index:idx_chat_pair;uniqueIndex:idx_chat_sender_client_message"`
	ReceiverID      uint    `gorm:"not null;index;index:idx_chat_pair;index:idx_chat_receiver_read"`
	ClientMessageID *string `gorm:"size:64;uniqueIndex:idx_chat_sender_client_message"`
	MessageType     string  `gorm:"size:20;not null;default:text;index"`
	Content         string  `gorm:"type:text;not null"`
	MediaURL        string  `gorm:"size:500"`
	MediaName       string  `gorm:"size:255"`
	SharedVideoID   *uint   `gorm:"index"`
	Read            bool    `gorm:"default:false;index;index:idx_chat_receiver_read"`
	Sender          User    `gorm:"foreignKey:SenderID"`
	Receiver        User    `gorm:"foreignKey:ReceiverID"`
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
	Status       string    `gorm:"size:20;not null;default:active;index"`
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
	Status       string     `gorm:"size:20;not null;default:pending;index"`
	PaidAt       *time.Time `gorm:"index"`
	Subscriber   User       `gorm:"foreignKey:SubscriberID"`
	Creator      User       `gorm:"foreignKey:CreatorID"`
}

// Video is an external projection used only to keep legacy response structs compiling.
// It is never migrated or queried by user-service.
type Video struct {
	gorm.Model
	Title    string
	CoverURL string
	Duration int
	Status   string
	AuthorID uint
	Author   User
}
