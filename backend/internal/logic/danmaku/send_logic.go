package danmakulogic

import (
	"context"
	"errors"
	"time"

	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"gorm.io/gorm"
)

type SendDanmakuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendDanmakuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendDanmakuLogic {
	return &SendDanmakuLogic{ctx: ctx, svcCtx: svcCtx}
}

type SendDanmakuReq struct {
	VideoID  uint   `json:"videoId"`
	Content  string `json:"content"`
	Time     int    `json:"time"`
	Color    string `json:"color,optional"`
	FontSize string `json:"fontSize,optional"`
	Type     string `json:"type,optional"`
}

type SendDanmakuResp struct {
	ID        uint      `json:"id"`
	VideoID   uint      `json:"videoId"`
	UserID    uint      `json:"userId"`
	Content   string    `json:"content"`
	Time      int       `json:"time"`
	Color     string    `json:"color"`
	FontSize  string    `json:"fontSize"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
}

func (l *SendDanmakuLogic) Send(req *SendDanmakuReq) (*SendDanmakuResp, error) {
	userID, ok := l.ctx.Value(middleware.CtxKeyUserID).(uint)
	if !ok || userID == 0 {
		return nil, errors.New("未登录")
	}

	if req.VideoID == 0 {
		return nil, errors.New("videoId 不能为空")
	}
	if req.Content == "" {
		return nil, errors.New("弹幕内容不能为空")
	}
	if req.Time < 0 {
		return nil, errors.New("弹幕时间不能为负数")
	}

	var video model.Video
	if err := l.svcCtx.DB.
		Where("id = ? AND status = ?", req.VideoID, "approved").
		First(&video).Error; err != nil {
		return nil, err
	}

	if req.Color == "" {
		req.Color = "#FFFFFF"
	}
	if req.FontSize == "" {
		req.FontSize = "medium"
	}
	if req.Type == "" {
		req.Type = "scroll"
	}

	danmaku := model.Danmaku{
		VideoID:  req.VideoID,
		UserID:   userID,
		Content:  req.Content,
		Time:     req.Time,
		Color:    req.Color,
		FontSize: req.FontSize,
		Type:     req.Type,
		Blocked:  false,
	}

	if err := l.svcCtx.DB.Create(&danmaku).Error; err != nil {
		return nil, err
	}

	if err := l.svcCtx.DB.Model(&video).
		UpdateColumn("danmaku_count", gorm.Expr("danmaku_count + ?", 1)).Error; err != nil {
		return nil, err
	}

	return &SendDanmakuResp{
		ID:        danmaku.ID,
		VideoID:   danmaku.VideoID,
		UserID:    danmaku.UserID,
		Content:   danmaku.Content,
		Time:      danmaku.Time,
		Color:     danmaku.Color,
		FontSize:  danmaku.FontSize,
		Type:      danmaku.Type,
		CreatedAt: danmaku.CreatedAt,
	}, nil
}