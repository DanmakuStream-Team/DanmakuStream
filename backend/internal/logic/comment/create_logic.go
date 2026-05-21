package commentlogic

import (
	"context"
	"errors"
	"time"

	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
)

type CreateCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{ctx: ctx, svcCtx: svcCtx}
}

type CreateCommentReq struct {
	VideoID  uint  `json:"videoId"`
	ParentID *uint `json:"parentId,optional"`
	Content  string `json:"content"`
}

type CommentAuthor struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

type CreateCommentResp struct {
	ID        uint           `json:"id"`
	VideoID   uint           `json:"videoId"`
	UserID    uint           `json:"userId"`
	ParentID  *uint          `json:"parentId"`
	Content   string         `json:"content"`
	LikeCount int64          `json:"likeCount"`
	Author    *CommentAuthor `json:"author"`
	CreatedAt time.Time      `json:"createdAt"`
}

func (l *CreateCommentLogic) Create(req *CreateCommentReq) (*CreateCommentResp, error) {
	userID, ok := l.ctx.Value(middleware.CtxKeyUserID).(uint)
	if !ok || userID == 0 {
		return nil, errors.New("未登录")
	}

	if req.VideoID == 0 {
		return nil, errors.New("videoId 不能为空")
	}
	if req.Content == "" {
		return nil, errors.New("评论内容不能为空")
	}

	var video model.Video
	if err := l.svcCtx.DB.
		Where("id = ? AND status = ?", req.VideoID, "approved").
		First(&video).Error; err != nil {
		return nil, err
	}

	if req.ParentID != nil {
		var parent model.Comment
		if err := l.svcCtx.DB.
			Where("id = ? AND video_id = ?", *req.ParentID, req.VideoID).
			First(&parent).Error; err != nil {
			return nil, err
		}
	}

	comment := model.Comment{
		VideoID:  req.VideoID,
		UserID:   userID,
		ParentID: req.ParentID,
		Content:  req.Content,
	}

	if err := l.svcCtx.DB.Create(&comment).Error; err != nil {
		return nil, err
	}

	var user model.User
	if err := l.svcCtx.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &CreateCommentResp{
		ID:        comment.ID,
		VideoID:   comment.VideoID,
		UserID:    comment.UserID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		LikeCount: comment.LikeCount,
		Author: &CommentAuthor{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Role:     user.Role,
		},
		CreatedAt: comment.CreatedAt,
	}, nil
}