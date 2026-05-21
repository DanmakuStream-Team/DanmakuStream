package userlogic

import (
	"context"
	"errors"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
)

type ProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileLogic {
	return &ProfileLogic{ctx: ctx, svcCtx: svcCtx}
}

type ProfileReq struct {
	ID uint `path:"id"`
}

type ProfileResp struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	Role        string    `json:"role"`
	FollowCount int64     `json:"followCount"`
	FanCount    int64     `json:"fanCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (l *ProfileLogic) Profile(req *ProfileReq) (*ProfileResp, error) {
	if req.ID == 0 {
		return nil, errors.New("invalid user id")
	}

	var user model.User
	if err := l.svcCtx.DB.First(&user, req.ID).Error; err != nil {
		return nil, err
	}

	return &ProfileResp{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		Bio:         user.Bio,
		Role:        user.Role,
		FollowCount: user.FollowCount,
		FanCount:    user.FanCount,
		CreatedAt:   user.CreatedAt,
	}, nil
}