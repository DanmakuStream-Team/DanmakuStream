package videologic

import (
	"context"
	"errors"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"gorm.io/gorm"
)

type DetailVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDetailVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailVideoLogic {
	return &DetailVideoLogic{ctx: ctx, svcCtx: svcCtx}
}

type VideoDetailReq struct {
	ID uint `uri:"id"`
}

type VideoDetailInfo struct {
	VideoInfo
	CommentCount int64 `json:"commentCount"`
}

func (l *DetailVideoLogic) Detail(req *VideoDetailReq) (*VideoDetailInfo, error) {
	if req.ID == 0 {
		return nil, errors.New("invalid video id")
	}

	if err := l.svcCtx.DB.Model(&model.Video{}).
		Where("id = ? AND status = ?", req.ID, "approved").
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error; err != nil {
		return nil, err
	}

	var video model.Video
	if err := l.svcCtx.DB.Preload("Author").
		Where("id = ? AND status = ?", req.ID, "approved").
		First(&video).Error; err != nil {
		return nil, err
	}

	var commentCount int64
	if err := l.svcCtx.DB.Model(&model.Comment{}).
		Where("video_id = ?", req.ID).
		Count(&commentCount).Error; err != nil {
		return nil, err
	}

	return &VideoDetailInfo{
		VideoInfo: VideoInfo{
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
			CreatedAt:    video.CreatedAt.Format("2006-01-02 15:04:05"),
			Author: &model.UserInfo{
				ID:       video.Author.ID,
				Username: video.Author.Username,
				Nickname: video.Author.Nickname,
				Avatar:   video.Author.Avatar,
				Role:     video.Author.Role,
			},
		},
		CommentCount: commentCount,
	}, nil
}
