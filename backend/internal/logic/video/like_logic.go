package videologic

import (
	"context"
	"errors"

	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"gorm.io/gorm"
)

type LikeVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeVideoLogic {
	return &LikeVideoLogic{ctx: ctx, svcCtx: svcCtx}
}

type VideoLikeReq struct {
	ID uint `path:"id"`
}

type VideoLikeResp struct {
	Liked     bool  `json:"liked"`
	LikeCount int64 `json:"likeCount"`
}

func (l *LikeVideoLogic) Like(req *VideoLikeReq) (*VideoLikeResp, error) {
	userID, ok := l.ctx.Value(middleware.CtxKeyUserID).(uint)
	if !ok || userID == 0 {
		return nil, errors.New("未登录")
	}

	var video model.Video
	if err := l.svcCtx.DB.First(&video, req.ID).Error; err != nil {
		return nil, err
	}

	var like model.Like
	err := l.svcCtx.DB.Where("user_id = ? AND video_id = ?", userID, req.ID).First(&like).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := l.svcCtx.DB.Create(&model.Like{
			UserID:  userID,
			VideoID: req.ID,
		}).Error; err != nil {
			return nil, err
		}

		if err := l.svcCtx.DB.Model(&video).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
			return nil, err
		}

		video.LikeCount++
		return &VideoLikeResp{Liked: true, LikeCount: video.LikeCount}, nil
	}

	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.DB.Delete(&like).Error; err != nil {
		return nil, err
	}

	if err := l.svcCtx.DB.Model(&video).
		UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
		return nil, err
	}

	video.LikeCount--
	return &VideoLikeResp{Liked: false, LikeCount: video.LikeCount}, nil
}