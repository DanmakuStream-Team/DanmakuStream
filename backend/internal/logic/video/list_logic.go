package videologic

import (
	"context"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
)

type ListVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVideoLogic {
	return &ListVideoLogic{ctx: ctx, svcCtx: svcCtx}
}

type VideoListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Keyword  string `form:"keyword,optional"`
	Tag      string `form:"tag,optional"`
}

type PageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type VideoInfo struct {
	ID           uint            `json:"id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	CoverURL     string          `json:"coverUrl"`
	VideoURL     string          `json:"videoUrl"`
	Duration     int             `json:"duration"`
	ViewCount    int64           `json:"viewCount"`
	LikeCount    int64           `json:"likeCount"`
	CollectCount int64           `json:"collectCount"`
	DanmakuCount int64           `json:"danmakuCount"`
	Tags         string          `json:"tags"`
	CreatedAt    time.Time       `json:"createdAt"`
	Author       *model.UserInfo `json:"author"`
}

func (l *ListVideoLogic) List(req *VideoListReq) (*PageResult[VideoInfo], error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	db := l.svcCtx.DB.Model(&model.Video{}).
		Preload("Author").
		Where("status = ?", "approved")

	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		db = db.Where("title LIKE ? OR description LIKE ?", like, like)
	}

	if req.Tag != "" {
		db = db.Where("FIND_IN_SET(?, tags)", req.Tag)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var videos []model.Video
	if err := db.Order("created_at DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&videos).Error; err != nil {
		return nil, err
	}

	list := make([]VideoInfo, 0, len(videos))
	for _, video := range videos {
		list = append(list, VideoInfo{
			ID:           video.ID,
			Title:        video.Title,
			Description:  video.Description,
			CoverURL:     video.CoverURL,
			VideoURL:     video.VideoURL,
			Duration:     video.Duration,
			ViewCount:    video.ViewCount,
			LikeCount:    video.LikeCount,
			CollectCount: video.CollectCount,
			DanmakuCount: video.DanmakuCount,
			Tags:         video.Tags,
			CreatedAt:    video.CreatedAt,
			Author: &model.UserInfo{
				ID:       video.Author.ID,
				Username: video.Author.Username,
				Nickname: video.Author.Nickname,
				Avatar:   video.Author.Avatar,
				Role:     video.Author.Role,
			},
		})
	}

	return &PageResult[VideoInfo]{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
