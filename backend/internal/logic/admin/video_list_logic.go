package adminlogic

import (
	"context"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
)

type VideoListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVideoListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoListLogic {
	return &VideoListLogic{ctx: ctx, svcCtx: svcCtx}
}

type AdminVideoListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Status   string `form:"status,optional"`
}

type PageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type AdminVideoInfo struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	CoverURL     string    `json:"coverUrl"`
	VideoURL     string    `json:"videoUrl"`
	Duration     int       `json:"duration"`
	ViewCount    int64     `json:"viewCount"`
	LikeCount    int64     `json:"likeCount"`
	CollectCount int64     `json:"collectCount"`
	DanmakuCount int64     `json:"danmakuCount"`
	Status       string    `json:"status"`
	AuthorID     uint      `json:"authorId"`
	Tags         string    `json:"tags"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (l *VideoListLogic) List(req *AdminVideoListReq) (*PageResult[AdminVideoInfo], error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	db := l.svcCtx.DB.Model(&model.Video{})

	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var videos []model.Video
	if err := db.
		Order("created_at DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&videos).Error; err != nil {
		return nil, err
	}

	list := make([]AdminVideoInfo, 0, len(videos))
	for _, v := range videos {
		list = append(list, AdminVideoInfo{
			ID:           v.ID,
			Title:        v.Title,
			Description:  v.Description,
			CoverURL:     v.CoverURL,
			VideoURL:     v.VideoURL,
			Duration:     v.Duration,
			ViewCount:    v.ViewCount,
			LikeCount:    v.LikeCount,
			CollectCount: v.CollectCount,
			DanmakuCount: v.DanmakuCount,
			Status:       v.Status,
			AuthorID:     v.AuthorID,
			Tags:         v.Tags,
			CreatedAt:    v.CreatedAt,
		})
	}

	return &PageResult[AdminVideoInfo]{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}