package danmakulogic

import (
	"context"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
)

type ListDanmakuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListDanmakuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDanmakuLogic {
	return &ListDanmakuLogic{ctx: ctx, svcCtx: svcCtx}
}

type DanmakuListReq struct {
	VideoID uint `path:"videoId"`
}

type DanmakuInfo struct {
	ID        uint   `json:"id"`
	VideoID   uint   `json:"videoId"`
	UserID    uint   `json:"userId"`
	Content   string `json:"content"`
	Time      int    `json:"time"`
	Color     string `json:"color"`
	FontSize  string `json:"fontSize"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
}

func (l *ListDanmakuLogic) List(req *DanmakuListReq) ([]DanmakuInfo, error) {
	var danmakus []model.Danmaku

	err := l.svcCtx.DB.
		Where("video_id = ? AND blocked = ?", req.VideoID, false).
		Order("time ASC").
		Find(&danmakus).Error

	if err != nil {
		return nil, err
	}

	list := make([]DanmakuInfo, 0, len(danmakus))
	for _, d := range danmakus {
		list = append(list, DanmakuInfo{
			ID:        d.ID,
			VideoID:   d.VideoID,
			UserID:    d.UserID,
			Content:   d.Content,
			Time:      d.Time,
			Color:     d.Color,
			FontSize:  d.FontSize,
			Type:      d.Type,
			CreatedAt: d.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return list, nil
}