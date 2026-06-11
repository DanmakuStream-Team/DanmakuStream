package collection

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"danmakustream/backend/internal/handler/response"
	videologic "danmakustream/backend/internal/logic/video"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createCollectionReq struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	CoverURL    string `json:"coverUrl"`
}

type addCollectionVideoReq struct {
	VideoID uint `json:"videoId" binding:"required"`
	Sort    int  `json:"sort"`
}

type addCollaboratorReq struct {
	UserID uint `json:"userId" binding:"required"`
}

type collectionInfo struct {
	ID          uint                   `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	CoverURL    string                 `json:"coverUrl"`
	Owner       *model.UserInfo        `json:"owner"`
	Videos      []videologic.VideoInfo `json:"videos,omitempty"`
	CreatedAt   string                 `json:"createdAt"`
}

func CreateHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createCollectionReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		title := strings.TrimSpace(req.Title)
		if title == "" {
			response.Fail(c, http.StatusBadRequest, "合集标题不能为空")
			return
		}

		collection := model.VideoCollection{
			Title:       title,
			Description: strings.TrimSpace(req.Description),
			CoverURL:    strings.TrimSpace(req.CoverURL),
			OwnerID:     c.GetUint(middleware.CtxKeyUserID),
		}
		if err := svcCtx.DB.Create(&collection).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "合集创建失败")
			return
		}

		info, err := loadCollectionInfo(svcCtx, collection.ID, false)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "合集加载失败")
			return
		}
		response.Ok(c, info)
	}
}

func MineHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		var collections []model.VideoCollection
		if err := svcCtx.DB.Preload("Owner").
			Where("owner_id = ?", userID).
			Order("created_at DESC").
			Find(&collections).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "合集加载失败")
			return
		}

		list := make([]collectionInfo, 0, len(collections))
		for _, collection := range collections {
			list = append(list, toCollectionInfo(collection, nil))
		}
		response.Ok(c, list)
	}
}

func DetailHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID(c, "id")
		if !ok {
			return
		}

		info, err := loadCollectionInfo(svcCtx, id, true)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, "合集不存在")
				return
			}
			response.Fail(c, http.StatusInternalServerError, "合集加载失败")
			return
		}
		response.Ok(c, info)
	}
}

func VideoCollectionsHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, ok := parseID(c, "id")
		if !ok {
			return
		}

		var items []model.VideoCollectionItem
		if err := svcCtx.DB.Preload("Collection.Owner").
			Joins("JOIN video_collections ON video_collections.id = video_collection_items.collection_id AND video_collections.deleted_at IS NULL").
			Where("video_collection_items.video_id = ?", videoID).
			Order("video_collections.created_at DESC").
			Find(&items).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "合集加载失败")
			return
		}

		list := make([]collectionInfo, 0, len(items))
		for _, item := range items {
			list = append(list, toCollectionInfo(item.Collection, nil))
		}
		response.Ok(c, list)
	}
}

func AddVideoHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		collectionID, ok := parseID(c, "id")
		if !ok {
			return
		}

		var req addCollectionVideoReq
		if err := c.ShouldBindJSON(&req); err != nil || req.VideoID == 0 {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		if !canManageCollection(svcCtx, collectionID, userID) {
			response.Fail(c, http.StatusForbidden, "无权管理该合集")
			return
		}
		if !canUseVideo(svcCtx, req.VideoID, userID) {
			response.Fail(c, http.StatusForbidden, "只能添加自己或共创的视频")
			return
		}

		item := model.VideoCollectionItem{
			CollectionID: collectionID,
			VideoID:      req.VideoID,
			Sort:         req.Sort,
		}
		if err := svcCtx.DB.Where("collection_id = ? AND video_id = ?", collectionID, req.VideoID).
			FirstOrCreate(&item).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "添加视频失败")
			return
		}

		response.Ok(c, gin.H{"id": item.ID})
	}
}

func RemoveVideoHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		collectionID, ok := parseID(c, "id")
		if !ok {
			return
		}
		videoID, ok := parseID(c, "videoId")
		if !ok {
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		if !canManageCollection(svcCtx, collectionID, userID) {
			response.Fail(c, http.StatusForbidden, "无权管理该合集")
			return
		}

		if err := svcCtx.DB.Where("collection_id = ? AND video_id = ?", collectionID, videoID).
			Delete(&model.VideoCollectionItem{}).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "移除视频失败")
			return
		}
		response.Ok(c, gin.H{"videoId": videoID})
	}
}

func AddCollaboratorHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, ok := parseID(c, "id")
		if !ok {
			return
		}

		var req addCollaboratorReq
		if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		if !isVideoAuthor(svcCtx, videoID, userID) {
			response.Fail(c, http.StatusForbidden, "只有视频作者可以添加共创")
			return
		}
		if req.UserID == userID {
			response.Fail(c, http.StatusBadRequest, "不能添加自己为共创")
			return
		}

		var user model.User
		if err := svcCtx.DB.Select("id").First(&user, req.UserID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "用户不存在")
			return
		}

		collaborator := model.VideoCollaborator{VideoID: videoID, UserID: req.UserID}
		if err := svcCtx.DB.Where("video_id = ? AND user_id = ?", videoID, req.UserID).
			FirstOrCreate(&collaborator).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "添加共创失败")
			return
		}

		response.Ok(c, gin.H{"videoId": videoID, "userId": req.UserID})
	}
}

func RemoveCollaboratorHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, ok := parseID(c, "id")
		if !ok {
			return
		}
		targetUserID, ok := parseID(c, "userId")
		if !ok {
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		if !isVideoAuthor(svcCtx, videoID, userID) {
			response.Fail(c, http.StatusForbidden, "只有视频作者可以移除共创")
			return
		}

		if err := svcCtx.DB.Where("video_id = ? AND user_id = ?", videoID, targetUserID).
			Delete(&model.VideoCollaborator{}).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "移除共创失败")
			return
		}
		response.Ok(c, gin.H{"videoId": videoID, "userId": targetUserID})
	}
}

func parseID(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的 ID")
		return 0, false
	}
	return uint(id), true
}

func loadCollectionInfo(svcCtx *svc.ServiceContext, id uint, withVideos bool) (*collectionInfo, error) {
	var collection model.VideoCollection
	if err := svcCtx.DB.Preload("Owner").First(&collection, id).Error; err != nil {
		return nil, err
	}

	var videos []videologic.VideoInfo
	if withVideos {
		var items []model.VideoCollectionItem
		if err := svcCtx.DB.Preload("Video.Author").
			Joins("JOIN videos ON videos.id = video_collection_items.video_id AND videos.deleted_at IS NULL").
			Where("video_collection_items.collection_id = ? AND videos.status = ?", id, "approved").
			Order("video_collection_items.sort ASC, video_collection_items.created_at ASC").
			Find(&items).Error; err != nil {
			return nil, err
		}

		videos = make([]videologic.VideoInfo, 0, len(items))
		for _, item := range items {
			videos = append(videos, toVideoInfo(item.Video))
		}
	}

	info := toCollectionInfo(collection, videos)
	return &info, nil
}

func toCollectionInfo(collection model.VideoCollection, videos []videologic.VideoInfo) collectionInfo {
	return collectionInfo{
		ID:          collection.ID,
		Title:       collection.Title,
		Description: collection.Description,
		CoverURL:    collection.CoverURL,
		Owner: &model.UserInfo{
			ID:       collection.Owner.ID,
			Username: collection.Owner.Username,
			Nickname: collection.Owner.Nickname,
			Avatar:   collection.Owner.Avatar,
			Role:     collection.Owner.Role,
		},
		Videos:    videos,
		CreatedAt: collection.CreatedAt.Format("2006-01-02 15:04:05"),
	}
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
		Category:     video.Category,
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

func canManageCollection(svcCtx *svc.ServiceContext, collectionID uint, userID uint) bool {
	var count int64
	svcCtx.DB.Model(&model.VideoCollection{}).
		Where("id = ? AND owner_id = ?", collectionID, userID).
		Count(&count)
	return count > 0
}

func canUseVideo(svcCtx *svc.ServiceContext, videoID uint, userID uint) bool {
	if isVideoAuthor(svcCtx, videoID, userID) {
		return true
	}

	var count int64
	svcCtx.DB.Model(&model.VideoCollaborator{}).
		Where("video_id = ? AND user_id = ?", videoID, userID).
		Count(&count)
	return count > 0
}

func isVideoAuthor(svcCtx *svc.ServiceContext, videoID uint, userID uint) bool {
	var count int64
	svcCtx.DB.Model(&model.Video{}).
		Where("id = ? AND author_id = ?", videoID, userID).
		Count(&count)
	return count > 0
}
