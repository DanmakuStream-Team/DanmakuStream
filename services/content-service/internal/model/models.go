package model

import (
	"time"

	"gorm.io/gorm"
)

type Video struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Title           string         `gorm:"size:200;not null" json:"title"`
	Description     string         `gorm:"type:text" json:"description"`
	CoverURL        string         `gorm:"size:500" json:"coverUrl"`
	VideoURL        string         `gorm:"size:500;not null" json:"videoUrl"`
	Duration        int            `gorm:"default:0" json:"duration"`
	ViewCount       int64          `gorm:"default:0" json:"viewCount"`
	LikeCount       int64          `gorm:"default:0" json:"likeCount"`
	CollectCount    int64          `gorm:"default:0" json:"collectCount"`
	DanmakuCount    int64          `gorm:"default:0" json:"danmakuCount"`
	Status          string         `gorm:"size:20;not null;default:pending;index" json:"status"`
	TranscodeStatus string         `gorm:"size:20;not null;default:ready" json:"transcodeStatus"`
	TranscodeError  string         `gorm:"size:500" json:"transcodeError,omitempty"`
	AuthorID        uint           `gorm:"not null;index" json:"authorId"`
	Tags            string         `gorm:"size:500" json:"tags"`
	Category        string         `gorm:"size:32;index" json:"category"`
}

type MediaAsset struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	OwnerID   uint           `gorm:"not null;index" json:"ownerId"`
	VideoID   *uint          `gorm:"index" json:"videoId,omitempty"`
	Kind      string         `gorm:"size:20;not null" json:"kind"`
	Path      string         `gorm:"size:500;not null;uniqueIndex" json:"url"`
	MimeType  string         `gorm:"size:100" json:"mimeType"`
	Size      int64          `json:"size"`
}

type VideoCollaborator struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	VideoID   uint           `gorm:"not null;uniqueIndex:idx_video_collaborator" json:"videoId"`
	UserID    uint           `gorm:"not null;uniqueIndex:idx_video_collaborator" json:"userId"`
}

type DynamicPost struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null;index" json:"userId"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Images    string         `gorm:"size:1000" json:"images"`
}

type SiteBanner struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Title     string         `gorm:"size:120;not null" json:"title"`
	ImageURL  string         `gorm:"size:500" json:"imageUrl"`
	Link      string         `gorm:"size:500" json:"link"`
	Enabled   bool           `gorm:"index" json:"enabled"`
	Sort      int            `gorm:"default:0" json:"sort"`
}

type SiteAnnouncement struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Content   string         `gorm:"size:500;not null" json:"content"`
	Enabled   bool           `gorm:"index" json:"enabled"`
	StartedAt *time.Time     `json:"startedAt"`
	EndedAt   *time.Time     `json:"endedAt"`
}

type CreatorDailyStat struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	CreatorID    uint           `gorm:"not null;uniqueIndex:idx_creator_daily_stat" json:"creatorId"`
	Date         string         `gorm:"size:10;not null;uniqueIndex:idx_creator_daily_stat" json:"date"`
	ViewDelta    int64          `json:"viewDelta"`
	CollectDelta int64          `json:"collectDelta"`
}

type VideoDailyStat struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	CreatorID    uint           `gorm:"not null;index" json:"creatorId"`
	VideoID      uint           `gorm:"not null;uniqueIndex:idx_video_daily_stat" json:"videoId"`
	Date         string         `gorm:"size:10;not null;uniqueIndex:idx_video_daily_stat" json:"date"`
	ViewDelta    int64          `json:"viewDelta"`
	CollectDelta int64          `json:"collectDelta"`
}

func ContentModels() []any {
	return []any{&Video{}, &MediaAsset{}, &VideoCollaborator{}, &DynamicPost{}, &SiteBanner{}, &SiteAnnouncement{}, &CreatorDailyStat{}, &VideoDailyStat{}}
}
