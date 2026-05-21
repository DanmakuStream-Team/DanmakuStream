package videologic

import (
	"context"
	"errors"

	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"gorm.io/gorm"
)

type CollectVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCollectVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectVideoLogic {
	return &CollectVideoLogic{ctx: ctx, svcCtx: svcCtx}
}

type VideoCollectReq struct {
	ID uint `path:"id"`
}

type VideoCollectResp struct {
	Collected    bool  `json:"collected"`
	CollectCount int64 `json:"collectCount"`
}

func (l *CollectVideoLogic) Collect(req *VideoCollectReq) (*VideoCollectResp, error) {
	userID, ok := l.ctx.Value(middleware.CtxKeyUserID).(uint)
	if !ok || userID == 0 {
		return nil, errors.New("未登录")
	}

	var video model.Video
	if err := l.svcCtx.DB.
		Where("id = ? AND status = ?", req.ID, "approved").
		First(&video).Error; err != nil {
		return nil, err
	}
	var collect model.Collect
	err := l.svcCtx.DB.Where("user_id = ? AND video_id = ?", userID, req.ID).First(&collect).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := l.svcCtx.DB.Create(&model.Collect{
			UserID:  userID,
			VideoID: req.ID,
		}).Error; err != nil {
			return nil, err
		}

		if err := l.svcCtx.DB.Model(&video).
			UpdateColumn("collect_count", gorm.Expr("collect_count + ?", 1)).Error; err != nil {
			return nil, err
		}

		video.CollectCount++
		return &VideoCollectResp{Collected: true, CollectCount: video.CollectCount}, nil
	}

	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.DB.Unscoped().Delete(&collect).Error; err != nil {
		return nil, err
	}

	if err := l.svcCtx.DB.Model(&video).
		UpdateColumn("collect_count", gorm.Expr("collect_count - ?", 1)).Error; err != nil {
		return nil, err
	}

	video.CollectCount--
	return &VideoCollectResp{Collected: false, CollectCount: video.CollectCount}, nil
}