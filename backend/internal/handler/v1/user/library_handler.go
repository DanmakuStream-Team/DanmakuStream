package user

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"danmakustream/backend/internal/handler/response"
	videologic "danmakustream/backend/internal/logic/video"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type updateHistoryReq struct {
	Position int `json:"position"`
}

type libraryRecord struct {
	Video    videologic.VideoInfo `json:"video"`
	SavedAt  string               `json:"savedAt"`
	Progress int                  `json:"progress"`
	Position int                  `json:"position"`
}

func HistoryListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		page, pageSize := libraryPage(c)
		db := svcCtx.DB.Model(&model.WatchHistory{}).
			Joins("JOIN videos ON videos.id = watch_histories.video_id AND videos.deleted_at IS NULL AND videos.status = ?", "approved").
			Where("watch_histories.user_id = ?", userID)

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "历史记录加载失败")
			return
		}

		histories := make([]model.WatchHistory, 0, pageSize)
		if err := db.Preload("Video.Author").
			Order("watch_histories.updated_at DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&histories).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "历史记录加载失败")
			return
		}

		list := make([]libraryRecord, 0, len(histories))
		for _, history := range histories {
			list = append(list, toLibraryRecord(history.Video, history.UpdatedAt, history.Position))
		}
		response.Ok(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
	}
}

func HistoryDetailHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		videoID, ok := libraryVideoID(c)
		if !ok {
			return
		}

		var history model.WatchHistory
		err := svcCtx.DB.Preload("Video.Author").
			Joins("JOIN videos ON videos.id = watch_histories.video_id AND videos.deleted_at IS NULL AND videos.status = ?", "approved").
			Where("watch_histories.user_id = ? AND watch_histories.video_id = ?", userID, videoID).
			First(&history).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "暂无观看记录")
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "历史记录加载失败")
			return
		}
		response.Ok(c, toLibraryRecord(history.Video, history.UpdatedAt, history.Position))
	}
}

func SaveHistoryHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		videoID, ok := libraryVideoID(c)
		if !ok {
			return
		}
		var req updateHistoryReq
		if err := c.ShouldBindJSON(&req); err != nil || req.Position < 0 {
			response.Fail(c, http.StatusBadRequest, "播放位置无效")
			return
		}

		var video model.Video
		if err := svcCtx.DB.Select("id", "duration").Where("id = ? AND status = ?", videoID, "approved").First(&video).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}
		if video.Duration > 0 && req.Position > video.Duration {
			req.Position = video.Duration
		}

		history := model.WatchHistory{UserID: userID, VideoID: uint(videoID), Position: req.Position}
		if err := svcCtx.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "video_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"position":   req.Position,
				"updated_at": time.Now(),
			}),
		}).Create(&history).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "观看进度保存失败")
			return
		}
		response.Ok(c, gin.H{"videoId": videoID, "position": req.Position})
	}
}

func DeleteHistoryHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		videoID, ok := libraryVideoID(c)
		if !ok {
			return
		}
		if err := svcCtx.DB.Unscoped().Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&model.WatchHistory{}).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "历史记录删除失败")
			return
		}
		response.Ok(c, gin.H{"videoId": videoID})
	}
}

func ClearHistoryHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		if err := svcCtx.DB.Unscoped().Where("user_id = ?", userID).Delete(&model.WatchHistory{}).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "历史记录清空失败")
			return
		}
		response.Ok(c, gin.H{"cleared": true})
	}
}

func WatchLaterListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		page, pageSize := libraryPage(c)
		db := svcCtx.DB.Model(&model.WatchLater{}).
			Joins("JOIN videos ON videos.id = watch_laters.video_id AND videos.deleted_at IS NULL AND videos.status = ?", "approved").
			Where("watch_laters.user_id = ?", userID)

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "稍后再看加载失败")
			return
		}
		items := make([]model.WatchLater, 0, pageSize)
		if err := db.Preload("Video.Author").Order("watch_laters.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "稍后再看加载失败")
			return
		}

		list := make([]libraryRecord, 0, len(items))
		for _, item := range items {
			list = append(list, toLibraryRecord(item.Video, item.CreatedAt, 0))
		}
		response.Ok(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
	}
}

func WatchLaterStatusHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		videoID, ok := libraryVideoID(c)
		if !ok {
			return
		}
		var count int64
		if err := svcCtx.DB.Model(&model.WatchLater{}).Where("user_id = ? AND video_id = ?", userID, videoID).Count(&count).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "稍后再看状态加载失败")
			return
		}
		response.Ok(c, gin.H{"saved": count > 0})
	}
}

func ToggleWatchLaterHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		videoID, ok := libraryVideoID(c)
		if !ok {
			return
		}
		var video model.Video
		if err := svcCtx.DB.Select("id").Where("id = ? AND status = ?", videoID, "approved").First(&video).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}

		var item model.WatchLater
		err := svcCtx.DB.Where("user_id = ? AND video_id = ?", userID, videoID).First(&item).Error
		if err == nil {
			if err := svcCtx.DB.Unscoped().Delete(&item).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "稍后再看操作失败")
				return
			}
			response.Ok(c, gin.H{"saved": false})
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusInternalServerError, "稍后再看操作失败")
			return
		}
		if err := svcCtx.DB.Create(&model.WatchLater{UserID: userID, VideoID: uint(videoID)}).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "稍后再看操作失败")
			return
		}
		response.Ok(c, gin.H{"saved": true})
	}
}

func DeleteWatchLaterHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		videoID, ok := libraryVideoID(c)
		if !ok {
			return
		}
		if err := svcCtx.DB.Unscoped().Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&model.WatchLater{}).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "稍后再看移除失败")
			return
		}
		response.Ok(c, gin.H{"videoId": videoID})
	}
}

func ClearWatchLaterHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		if err := svcCtx.DB.Unscoped().Where("user_id = ?", userID).Delete(&model.WatchLater{}).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "稍后再看清空失败")
			return
		}
		response.Ok(c, gin.H{"cleared": true})
	}
}

func libraryVideoID(c *gin.Context) (uint64, bool) {
	videoID, err := strconv.ParseUint(c.Param("videoId"), 10, 64)
	if err != nil || videoID == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
		return 0, false
	}
	return videoID, true
}

func libraryPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func toLibraryRecord(video model.Video, savedAt time.Time, position int) libraryRecord {
	progress := 0
	if video.Duration > 0 {
		progress = position * 100 / video.Duration
		if progress > 100 {
			progress = 100
		}
	}
	return libraryRecord{
		Video: videologic.VideoInfo{
			ID: video.ID, Title: video.Title, Description: video.Description,
			CoverURL: video.CoverURL, VideoURL: video.VideoURL, Duration: video.Duration,
			ViewCount: video.ViewCount, LikeCount: video.LikeCount, CollectCount: video.CollectCount,
			DanmakuCount: video.DanmakuCount, Status: video.Status,
			TranscodeStatus: videologic.EffectiveTranscodeStatus(video), TranscodeError: video.TranscodeError, Tags: video.Tags,
			Category: video.Category, CreatedAt: video.CreatedAt.Format("2006-01-02 15:04:05"),
			Author: &model.UserInfo{ID: video.Author.ID, Username: video.Author.Username, Nickname: video.Author.Nickname, Avatar: video.Author.Avatar, Role: video.Author.Role},
		},
		SavedAt: savedAt.Format("2006-01-02 15:04:05"), Progress: progress, Position: position,
	}
}
