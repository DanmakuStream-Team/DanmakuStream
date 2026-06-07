package comment

import (
	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type createCommentReq struct {
	VideoID  uint   `json:"videoId" binding:"required"`
	Content  string `json:"content" binding:"required,max=500"`
	ParentID *uint  `json:"parentId"`
}

type listCommentReq struct {
	Sort string `form:"sort"`
}

type commentItem struct {
	ID        uint            `json:"id"`
	VideoID   uint            `json:"videoId"`
	UserID    uint            `json:"userId"`
	Content   string          `json:"content"`
	LikeCount int64           `json:"likeCount"`
	Liked     bool            `json:"liked"`
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

		var req listCommentReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		orderExpr, err := commentSortExpr(req.Sort)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, err.Error())
			return
		}

		var comments []model.Comment
		err = svcCtx.DB.
			Preload("User").
			Where("video_id = ?", videoID).
			Order(orderExpr).
			Find(&comments).Error
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "评论加载失败")
			return
		}

		likedMap := map[uint]bool{}

		currentUserID := getOptionalUserID(c, svcCtx)
		if currentUserID != 0 && len(comments) > 0 {
			commentIDs := make([]uint, 0, len(comments))
			for _, comment := range comments {
				commentIDs = append(commentIDs, comment.ID)
			}

			var likes []model.CommentLike
			if err := svcCtx.DB.
				Where("user_id = ? AND comment_id IN ?", currentUserID, commentIDs).
				Find(&likes).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "评论点赞状态加载失败")
				return
			}

			for _, like := range likes {
				likedMap[like.CommentID] = true
			}
		}

		response.Ok(c, buildCommentTree(comments, likedMap))
	}
}

func commentSortExpr(sort string) (string, error) {
	switch sort {
	case "":
		return "created_at ASC", nil
	case "date":
		return "created_at DESC", nil
	case "like":
		return "like_count DESC, created_at DESC", nil
	default:
		return "", errors.New("无效的评论排序方式")
	}
}

func buildCommentTree(comments []model.Comment, likedMap map[uint]bool) []*commentItem {
	itemMap := make(map[uint]*commentItem, len(comments))
	roots := make([]*commentItem, 0)

	for _, comment := range comments {
		itemMap[comment.ID] = toCommentItem(comment, likedMap[comment.ID])
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

func toCommentItem(comment model.Comment, liked bool) *commentItem {
	return &commentItem{
		ID:        comment.ID,
		VideoID:   comment.VideoID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		LikeCount: comment.LikeCount,
		Liked:     liked,
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

		response.Ok(c, toCommentItem(comment, false))
	}
}

func DeleteHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)

		commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || commentID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的评论 ID")
			return
		}

		var comment model.Comment
		if err := svcCtx.DB.First(&comment, commentID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "评论不存在")
			return
		}

		if comment.UserID != userID && role != "admin" {
			response.Fail(c, http.StatusForbidden, "无权删除该评论")
			return
		}

		if err := svcCtx.DB.Delete(&comment).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "评论删除失败")
			return
		}

		response.Ok(c, gin.H{
			"id": comment.ID,
		})
	}
}

func LikeHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || commentID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的评论 ID")
			return
		}

		var comment model.Comment
		if err := svcCtx.DB.First(&comment, commentID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "评论不存在")
			return
		}

		var liked bool
		var likeCount int64

		err = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var commentLike model.CommentLike
			err := tx.Where("user_id = ? AND comment_id = ?", userID, commentID).
				First(&commentLike).Error

			if err == nil {
				if err := tx.Unscoped().Delete(&commentLike).Error; err != nil {
					return err
				}
				liked = false

				if err := tx.Model(&model.Comment{}).
					Where("id = ? AND like_count > 0", commentID).
					UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
					return err
				}
			} else {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				if err := tx.Create(&model.CommentLike{
					UserID:    userID,
					CommentID: uint(commentID),
				}).Error; err != nil {
					return err
				}

				liked = true

				if err := tx.Model(&model.Comment{}).
					Where("id = ?", commentID).
					UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
					return err
				}
			}

			return tx.Model(&model.Comment{}).
				Where("id = ?", commentID).
				Select("like_count").
				Scan(&likeCount).Error
		})

		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "评论点赞操作失败")
			return
		}

		response.Ok(c, gin.H{
			"liked":     liked,
			"likeCount": likeCount,
		})
	}
}

func getOptionalUserID(c *gin.Context, svcCtx *svc.ServiceContext) uint {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(svcCtx.Config.Auth.AccessSecret), nil
	})
	if err != nil || !token.Valid {
		return 0
	}

	return claims.UserID
}
