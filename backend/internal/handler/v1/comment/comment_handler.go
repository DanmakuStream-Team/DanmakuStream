package comment

import (
	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type createCommentReq struct {
	VideoID  uint   `json:"videoId" binding:"required"`
	Content  string `json:"content" binding:"required,max=500"`
	ParentID *uint  `json:"parentId"`
}

type commentItem struct {
	ID        uint            `json:"id"`
	VideoID   uint            `json:"videoId"`
	UserID    uint            `json:"userId"`
	Content   string          `json:"content"`
	LikeCount int64           `json:"likeCount"`
	Author    *model.UserInfo `json:"author"`
	Replies   []*commentItem  `json:"replies"`
	CreatedAt string          `json:"createdAt"`
}

func ListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.ParseUint(c.Param("videoId"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var comments []model.Comment
		err = svcCtx.DB.
			Preload("User").
			Where("video_id = ?", videoID).
			Order("created_at ASC").
			Find(&comments).Error
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "评论加载失败")
			return
		}

		response.Ok(c, buildCommentTree(comments))
	}
}

func buildCommentTree(comments []model.Comment) []*commentItem {
	itemMap := make(map[uint]*commentItem, len(comments))
	roots := make([]*commentItem, 0)

	for _, comment := range comments {
		itemMap[comment.ID] = toCommentItem(comment)
	}

	for _, comment := range comments {
		item := itemMap[comment.ID]

		if comment.ParentID == nil {
			roots = append(roots, item)
			continue
		}

		parent := itemMap[*comment.ParentID]
		if parent == nil {
			roots = append(roots, item)
			continue
		}

		parent.Replies = append(parent.Replies, item)
	}

	return roots
}

func toCommentItem(comment model.Comment) *commentItem {
	return &commentItem{
		ID:        comment.ID,
		VideoID:   comment.VideoID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		LikeCount: comment.LikeCount,
		CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
		Replies:   []*commentItem{},
		Author: &model.UserInfo{
			ID:       comment.User.ID,
			Username: comment.User.Username,
			Nickname: comment.User.Nickname,
			Avatar:   comment.User.Avatar,
			Role:     comment.User.Role,
		},
	}
}
func CreateHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createCommentReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		content := strings.TrimSpace(req.Content)
		if content == "" {
			response.Fail(c, http.StatusBadRequest, "评论内容不能为空")
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		if userID == 0 {
			response.Fail(c, http.StatusUnauthorized, "未授权")
			return
		}

		var video model.Video
		if err := svcCtx.DB.Select("id", "status").First(&video, req.VideoID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}
		if video.Status != "approved" {
			response.Fail(c, http.StatusForbidden, "视频未通过审核")
			return
		}

		if req.ParentID != nil {
			var parent model.Comment
			if err := svcCtx.DB.Select("id", "video_id").First(&parent, *req.ParentID).Error; err != nil {
				response.Fail(c, http.StatusNotFound, "父评论不存在")
				return
			}
			if parent.VideoID != req.VideoID {
				response.Fail(c, http.StatusBadRequest, "父评论不属于当前视频")
				return
			}
		}

		comment := model.Comment{
			VideoID:  req.VideoID,
			UserID:   userID,
			ParentID: req.ParentID,
			Content:  content,
		}

		if err := svcCtx.DB.Create(&comment).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "评论发布失败")
			return
		}

		if err := svcCtx.DB.Preload("User").First(&comment, comment.ID).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "评论加载失败")
			return
		}

		response.Ok(c, toCommentItem(comment))
	}
}
