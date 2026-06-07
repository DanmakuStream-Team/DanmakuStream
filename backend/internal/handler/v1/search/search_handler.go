package search

import (
	"net/http"
	"strings"

	"danmakustream/backend/internal/handler/response"
	videologic "danmakustream/backend/internal/logic/video"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type searchReq struct {
	Keyword  string `form:"keyword"`
	Q        string `form:"q"`
	Type     string `form:"type"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type userSearchInfo struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	Role        string `json:"role"`
	FollowCount int64  `json:"followCount"`
	FanCount    int64  `json:"fanCount"`
	VideoCount  int64  `json:"videoCount"`
	CreatedAt   string `json:"createdAt"`
}

func Handler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req searchReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		keyword := strings.TrimSpace(req.Keyword)
		if keyword == "" {
			keyword = strings.TrimSpace(req.Q)
		}
		if keyword == "" {
			response.Fail(c, http.StatusBadRequest, "搜索关键词不能为空")
			return
		}

		page, pageSize := normalizePage(req.Page, req.PageSize)
		searchType := strings.ToLower(strings.TrimSpace(req.Type))
		if searchType == "" {
			searchType = "all"
		}
		if searchType != "all" && searchType != "video" && searchType != "user" {
			response.Fail(c, http.StatusBadRequest, "搜索类型无效")
			return
		}

		result := gin.H{
			"keyword":  keyword,
			"type":     searchType,
			"page":     page,
			"pageSize": pageSize,
		}

		if searchType == "all" || searchType == "video" {
			videos, err := searchVideos(svcCtx, keyword, page, pageSize)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, "视频搜索失败")
				return
			}
			result["videos"] = videos
		}

		if searchType == "all" || searchType == "user" {
			users, err := searchUsers(svcCtx, keyword, page, pageSize)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, "用户搜索失败")
				return
			}
			result["users"] = users
		}

		response.Ok(c, result)
	}
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func searchVideos(svcCtx *svc.ServiceContext, keyword string, page, pageSize int) (*videologic.PageResult[videologic.VideoInfo], error) {
	likeExpr := "%" + keyword + "%"
	prefixExpr := keyword + "%"

	db := svcCtx.DB.Model(&model.Video{}).
		Preload("Author").
		Where("status = ?", "approved").
		Where("title LIKE ? OR description LIKE ? OR tags LIKE ?", likeExpr, likeExpr, likeExpr)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	const hotScoreExpr = "(like_count * 5 + collect_count * 3 + danmaku_count * 2 + view_count) " +
		"/ POW(GREATEST(TIMESTAMPDIFF(HOUR, created_at, NOW()), 0) + 2, 1.2)"
	orderExpr := "CASE " +
		"WHEN title = ? THEN 0 " +
		"WHEN title LIKE ? THEN 1 " +
		"WHEN tags LIKE ? THEN 2 " +
		"WHEN title LIKE ? THEN 3 " +
		"ELSE 4 END ASC, " + hotScoreExpr + " DESC"

	var videos []model.Video
	if err := db.
		Order(gorm.Expr(orderExpr, keyword, prefixExpr, likeExpr, likeExpr)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&videos).Error; err != nil {
		return nil, err
	}

	list := make([]videologic.VideoInfo, 0, len(videos))
	for _, video := range videos {
		list = append(list, toVideoInfo(video))
	}

	return &videologic.PageResult[videologic.VideoInfo]{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func searchUsers(svcCtx *svc.ServiceContext, keyword string, page, pageSize int) (*videologic.PageResult[userSearchInfo], error) {
	likeExpr := "%" + keyword + "%"
	prefixExpr := keyword + "%"

	db := svcCtx.DB.Model(&model.User{}).
		Where("username LIKE ? OR nickname LIKE ? OR bio LIKE ?", likeExpr, likeExpr, likeExpr)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []model.User
	if err := db.
		Order(gorm.Expr(
			"CASE WHEN nickname = ? THEN 0 WHEN username = ? THEN 1 WHEN nickname LIKE ? THEN 2 WHEN username LIKE ? THEN 3 ELSE 4 END ASC, fan_count DESC, id DESC",
			keyword, keyword, prefixExpr, prefixExpr,
		)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, err
	}

	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	videoCounts := make(map[uint]int64, len(users))
	if len(userIDs) > 0 {
		type videoCountRow struct {
			AuthorID uint
			Count    int64
		}
		var rows []videoCountRow
		if err := svcCtx.DB.Model(&model.Video{}).
			Select("author_id, COUNT(*) AS count").
			Where("author_id IN ? AND status = ?", userIDs, "approved").
			Group("author_id").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			videoCounts[row.AuthorID] = row.Count
		}
	}

	list := make([]userSearchInfo, 0, len(users))
	for _, user := range users {
		list = append(list, userSearchInfo{
			ID:          user.ID,
			Username:    user.Username,
			Nickname:    user.Nickname,
			Avatar:      user.Avatar,
			Bio:         user.Bio,
			Role:        user.Role,
			FollowCount: user.FollowCount,
			FanCount:    user.FanCount,
			VideoCount:  videoCounts[user.ID],
			CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &videologic.PageResult[userSearchInfo]{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func toVideoInfo(video model.Video) videologic.VideoInfo {
	return videologic.VideoInfo{
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
		Status:       video.Status,
		Tags:         video.Tags,
		CreatedAt:    video.CreatedAt.Format("2006-01-02 15:04:05"),
		Author: &model.UserInfo{
			ID:       video.Author.ID,
			Username: video.Author.Username,
			Nickname: video.Author.Nickname,
			Avatar:   video.Author.Avatar,
			Role:     video.Author.Role,
		},
	}
}
