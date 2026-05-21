package userlogic

import (
	"context"
	"errors"

	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"gorm.io/gorm"
)

type FollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{ctx: ctx, svcCtx: svcCtx}
}

type FollowReq struct {
	ID uint `path:"id"`
}

type FollowResp struct {
	Followed    bool  `json:"followed"`
	FollowCount int64 `json:"followCount"`
	FanCount    int64 `json:"fanCount"`
}

func (l *FollowLogic) Follow(req *FollowReq) (*FollowResp, error) {
	userID, ok := l.ctx.Value(middleware.CtxKeyUserID).(uint)
	if !ok || userID == 0 {
		return nil, errors.New("未登录")
	}

	if req.ID == 0 {
		return nil, errors.New("invalid user id")
	}

	if userID == req.ID {
		return nil, errors.New("不能关注自己")
	}

	var target model.User
	if err := l.svcCtx.DB.First(&target, req.ID).Error; err != nil {
		return nil, err
	}

	var current model.User
	if err := l.svcCtx.DB.First(&current, userID).Error; err != nil {
		return nil, err
	}

	var follow model.Follow
	err := l.svcCtx.DB.
		Where("follower_id = ? AND followee_id = ?", userID, req.ID).
		First(&follow).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := l.svcCtx.DB.Create(&model.Follow{
			FollowerID: userID,
			FolloweeID: req.ID,
		}).Error; err != nil {
			return nil, err
		}

		if err := l.svcCtx.DB.Model(&current).
			UpdateColumn("follow_count", gorm.Expr("follow_count + ?", 1)).Error; err != nil {
			return nil, err
		}

		if err := l.svcCtx.DB.Model(&target).
			UpdateColumn("fan_count", gorm.Expr("fan_count + ?", 1)).Error; err != nil {
			return nil, err
		}

		current.FollowCount++
		target.FanCount++

		return &FollowResp{
			Followed:    true,
			FollowCount: current.FollowCount,
			FanCount:    target.FanCount,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.DB.Unscoped().Delete(&follow).Error; err != nil {
		return nil, err
	}

	if err := l.svcCtx.DB.Model(&current).
		UpdateColumn("follow_count", gorm.Expr("follow_count - ?", 1)).Error; err != nil {
		return nil, err
	}

	if err := l.svcCtx.DB.Model(&target).
		UpdateColumn("fan_count", gorm.Expr("fan_count - ?", 1)).Error; err != nil {
		return nil, err
	}

	current.FollowCount--
	target.FanCount--

	return &FollowResp{
		Followed:    false,
		FollowCount: current.FollowCount,
		FanCount:    target.FanCount,
	}, nil
}