package logic

import (
	"errors"
	"strings"
	"time"

	"danmakustream/content-service/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNotFound  = errors.New("resource not found")
	ErrForbidden = errors.New("resource ownership required")
	ErrConflict  = errors.New("resource state conflict")
	ErrInvalid   = errors.New("invalid query")
)

type Service struct {
	DB *gorm.DB
}

type Page[T any] struct {
	Items    []T   `json:"items"`
	List     []T   `json:"list"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}

type AuthorDTO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

type VideoDTO struct {
	model.Video
	Author       AuthorDTO `json:"author"`
	CommentCount int64     `json:"commentCount"`
}

func VideoView(video model.Video) VideoDTO {
	return VideoDTO{Video: video, Author: AuthorDTO{ID: video.AuthorID}}
}

type VideoListOptions struct {
	Page              int
	PageSize          int
	AuthorID          *uint
	IncludeUnapproved bool
	Keyword           string
	Tag               string
	Category          string
	Status            string
	Sort              string
}

func (s *Service) ListVideos(options VideoListOptions) (Page[VideoDTO], error) {
	query := s.DB.Model(&model.Video{})
	if !options.IncludeUnapproved {
		query = query.Where("status = ? AND transcode_status = ?", "approved", "ready")
	}
	if options.AuthorID != nil {
		query = query.Where("author_id = ?", *options.AuthorID)
	}
	if options.Status != "" {
		if options.Status != "pending" && options.Status != "approved" && options.Status != "rejected" {
			return Page[VideoDTO]{}, ErrInvalid
		}
		query = query.Where("status = ?", options.Status)
	}
	if options.Keyword != "" {
		pattern := "%" + options.Keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", pattern, pattern)
	}
	if options.Tag != "" {
		query = query.Where("tags = ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ?", options.Tag, options.Tag+",%", "%,"+options.Tag, "%,"+options.Tag+",%")
	}
	if options.Category != "" {
		query = query.Where("category = ?", options.Category)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return Page[VideoDTO]{}, err
	}
	var videos []model.Video
	order, err := videoOrder(options.Sort)
	if err != nil {
		return Page[VideoDTO]{}, err
	}
	if err := query.Order(order).Offset((options.Page - 1) * options.PageSize).Limit(options.PageSize).Find(&videos).Error; err != nil {
		return Page[VideoDTO]{}, err
	}
	items := make([]VideoDTO, 0, len(videos))
	for _, video := range videos {
		items = append(items, VideoView(video))
	}
	return Page[VideoDTO]{Items: items, List: items, Page: options.Page, PageSize: options.PageSize, Total: total}, nil
}

func videoOrder(sort string) (string, error) {
	switch sort {
	case "", "hot":
		return "(like_count * 5 + collect_count * 3 + danmaku_count * 2 + view_count) DESC, created_at DESC", nil
	case "date":
		return "created_at DESC", nil
	case "like":
		return "like_count DESC, created_at DESC", nil
	case "collect":
		return "collect_count DESC, created_at DESC", nil
	default:
		return "", ErrInvalid
	}
}

func (s *Service) RecordView(video *model.Video) error {
	date := time.Now().Format("2006-01-02")
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Video{}).Where("id = ?", video.ID).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
			return err
		}
		creatorStat := model.CreatorDailyStat{CreatorID: video.AuthorID, Date: date, ViewDelta: 1}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "creator_id"}, {Name: "date"}}, DoUpdates: clause.Assignments(map[string]any{"view_delta": gorm.Expr("view_delta + 1")})}).Create(&creatorStat).Error; err != nil {
			return err
		}
		videoStat := model.VideoDailyStat{CreatorID: video.AuthorID, VideoID: video.ID, Date: date, ViewDelta: 1}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "video_id"}, {Name: "date"}}, DoUpdates: clause.Assignments(map[string]any{"view_delta": gorm.Expr("view_delta + 1")})}).Create(&videoStat).Error; err != nil {
			return err
		}
		video.ViewCount++
		return nil
	})
}

func (s *Service) GetVideo(id uint, includeUnapproved bool) (model.Video, error) {
	var video model.Video
	query := s.DB
	if !includeUnapproved {
		query = query.Where("status = ? AND transcode_status = ?", "approved", "ready")
	}
	if err := query.First(&video, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Video{}, ErrNotFound
	} else if err != nil {
		return model.Video{}, err
	}
	return video, nil
}

func (s *Service) CreateVideo(video *model.Video) error {
	video.Status = "pending"
	if video.TranscodeStatus == "" {
		video.TranscodeStatus = "ready"
	}
	return s.DB.Create(video).Error
}

func (s *Service) CanEditVideo(videoID, userID uint) (bool, error) {
	var video model.Video
	if err := s.DB.Select("id", "author_id").First(&video, videoID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return false, ErrNotFound
	} else if err != nil {
		return false, err
	}
	if video.AuthorID == userID {
		return true, nil
	}
	var count int64
	err := s.DB.Model(&model.VideoCollaborator{}).Where("video_id = ? AND user_id = ?", videoID, userID).Count(&count).Error
	return count > 0, err
}

func (s *Service) UpdateVideo(videoID, userID uint, fields map[string]any) (model.Video, error) {
	allowed, err := s.CanEditVideo(videoID, userID)
	if err != nil {
		return model.Video{}, err
	}
	if !allowed {
		return model.Video{}, ErrForbidden
	}
	fields["status"] = "pending"
	if err := s.DB.Model(&model.Video{}).Where("id = ?", videoID).Updates(fields).Error; err != nil {
		return model.Video{}, err
	}
	return s.GetVideo(videoID, true)
}

func (s *Service) DeleteVideo(videoID, userID uint) error {
	var video model.Video
	if err := s.DB.First(&video, videoID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	if video.AuthorID != userID {
		return ErrForbidden
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("video_id = ?", videoID).Delete(&model.VideoCollaborator{}).Error; err != nil {
			return err
		}
		if err := tx.Where("video_id = ?", videoID).Delete(&model.MediaAsset{}).Error; err != nil {
			return err
		}
		return tx.Delete(&video).Error
	})
}

func (s *Service) ReviewVideo(videoID uint, status string) (model.Video, error) {
	if status != "approved" && status != "rejected" {
		return model.Video{}, ErrConflict
	}
	video, err := s.GetVideo(videoID, true)
	if err != nil {
		return model.Video{}, err
	}
	if video.Status != "pending" {
		return model.Video{}, ErrConflict
	}
	if status == "approved" && video.TranscodeStatus != "ready" {
		return model.Video{}, ErrConflict
	}
	result := s.DB.Model(&model.Video{}).Where("id = ? AND status = ?", videoID, "pending").Update("status", status)
	if result.Error != nil {
		return model.Video{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.Video{}, ErrConflict
	}
	video.Status = status
	return video, nil
}

func (s *Service) AddCollaborator(videoID, ownerID, collaboratorID uint) (model.VideoCollaborator, error) {
	video, err := s.GetVideo(videoID, true)
	if err != nil {
		return model.VideoCollaborator{}, err
	}
	if video.AuthorID != ownerID {
		return model.VideoCollaborator{}, ErrForbidden
	}
	if collaboratorID == 0 || collaboratorID == ownerID {
		return model.VideoCollaborator{}, ErrConflict
	}
	item := model.VideoCollaborator{VideoID: videoID, UserID: collaboratorID}
	if err := s.DB.Create(&item).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return model.VideoCollaborator{}, ErrConflict
		}
		return model.VideoCollaborator{}, err
	}
	return item, nil
}

func (s *Service) RemoveCollaborator(videoID, ownerID, collaboratorID uint) error {
	video, err := s.GetVideo(videoID, true)
	if err != nil {
		return err
	}
	if video.AuthorID != ownerID {
		return ErrForbidden
	}
	result := s.DB.Unscoped().Where("video_id = ? AND user_id = ?", videoID, collaboratorID).Delete(&model.VideoCollaborator{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

type DynamicDTO struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Content   string    `json:"content"`
	Images    string    `json:"images"`
	Author    AuthorDTO `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *Service) ListDynamics(page, pageSize int, userID *uint) (Page[DynamicDTO], error) {
	query := s.DB.Model(&model.DynamicPost{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return Page[DynamicDTO]{}, err
	}
	var posts []model.DynamicPost
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&posts).Error; err != nil {
		return Page[DynamicDTO]{}, err
	}
	items := make([]DynamicDTO, 0, len(posts))
	for _, post := range posts {
		items = append(items, DynamicDTO{ID: post.ID, UserID: post.UserID, Content: post.Content, Images: post.Images, Author: AuthorDTO{ID: post.UserID}, CreatedAt: post.CreatedAt})
	}
	return Page[DynamicDTO]{Items: items, List: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *Service) DeleteDynamic(id, userID uint, isAdmin bool) error {
	query := s.DB.Where("id = ?", id)
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}
	result := query.Delete(&model.DynamicPost{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := s.DB.Model(&model.DynamicPost{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrForbidden
		}
		return ErrNotFound
	}
	return nil
}

type AnalyticsPoint struct {
	Date        string `json:"date"`
	Views       int64  `json:"views"`
	Collects    int64  `json:"collects"`
	GrowthSpeed int64  `json:"growthSpeed"`
	Streams     int64  `json:"streams"`
}
type AnalyticsSummary struct {
	TotalViews        int64   `json:"totalViews"`
	TotalCollects     int64   `json:"totalCollects"`
	TotalStreams      int64   `json:"totalStreams"`
	RangeViews        int64   `json:"rangeViews"`
	RangeCollects     int64   `json:"rangeCollects"`
	RangeStreams      int64   `json:"rangeStreams"`
	AverageDailyViews float64 `json:"averageDailyViews"`
}
type Analytics struct {
	Days            int              `json:"days"`
	SelectedVideoID uint             `json:"selectedVideoId"`
	Summary         AnalyticsSummary `json:"summary"`
	Points          []AnalyticsPoint `json:"points"`
	TopVideos       []VideoDTO       `json:"topVideos"`
}

func (s *Service) CreatorAnalytics(userID uint, days int, videoID uint) (Analytics, error) {
	if days != 7 && days != 30 && days != 90 {
		return Analytics{}, ErrInvalid
	}
	videoScope := func() *gorm.DB {
		query := s.DB.Model(&model.Video{}).Where("author_id = ?", userID)
		if videoID > 0 {
			query = query.Where("id = ?", videoID)
		}
		return query
	}
	if videoID > 0 {
		var count int64
		if err := videoScope().Count(&count).Error; err != nil {
			return Analytics{}, err
		}
		if count == 0 {
			return Analytics{}, ErrNotFound
		}
	}
	var totals struct {
		Views    int64
		Collects int64
	}
	if err := videoScope().Select("COALESCE(SUM(view_count), 0) AS views, COALESCE(SUM(collect_count), 0) AS collects").Scan(&totals).Error; err != nil {
		return Analytics{}, err
	}
	startDate := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	stats := s.DB.Model(&model.CreatorDailyStat{}).Where("creator_id = ? AND date >= ?", userID, startDate)
	if videoID > 0 {
		stats = s.DB.Model(&model.VideoDailyStat{}).Where("creator_id = ? AND video_id = ? AND date >= ?", userID, videoID, startDate)
	}
	var rows []struct {
		Date     string
		Views    int64
		Collects int64
	}
	if err := stats.Select("date, COALESCE(SUM(view_delta), 0) AS views, COALESCE(SUM(collect_delta), 0) AS collects").Group("date").Order("date ASC").Scan(&rows).Error; err != nil {
		return Analytics{}, err
	}
	byDate := map[string]AnalyticsPoint{}
	for _, row := range rows {
		byDate[row.Date] = AnalyticsPoint{Date: row.Date, Views: row.Views, Collects: row.Collects, GrowthSpeed: row.Views + max64(row.Collects, 0)}
	}
	points := make([]AnalyticsPoint, 0, days)
	var rangeViews, rangeCollects int64
	for offset := days - 1; offset >= 0; offset-- {
		date := time.Now().AddDate(0, 0, -offset).Format("2006-01-02")
		point := byDate[date]
		point.Date = date
		points = append(points, point)
		rangeViews += point.Views
		rangeCollects += point.Collects
	}
	var videos []model.Video
	if err := videoScope().Order("view_count DESC, collect_count DESC, created_at DESC").Limit(5).Find(&videos).Error; err != nil {
		return Analytics{}, err
	}
	top := make([]VideoDTO, 0, len(videos))
	for _, video := range videos {
		top = append(top, VideoView(video))
	}
	return Analytics{Days: days, SelectedVideoID: videoID, Summary: AnalyticsSummary{TotalViews: totals.Views, TotalCollects: totals.Collects, RangeViews: rangeViews, RangeCollects: rangeCollects, AverageDailyViews: float64(rangeViews) / float64(days)}, Points: points, TopVideos: top}, nil
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
