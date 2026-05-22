package video

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"gorm.io/gorm"

	"danmakustream/backend/internal/handler/response"
	videologic "danmakustream/backend/internal/logic/video"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
)

type adminVideoListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
}

type updateVideoStatusReq struct {
	Status string `json:"status" binding:"required"`
}

func isValidVideoStatus(status string) bool {
	return status == "pending" || status == "approved" || status == "rejected"
}

func ListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req videologic.VideoListReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		l := videologic.NewListVideoLogic(c.Request.Context(), svcCtx)
		resp, err := l.List(&req)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.Ok(c, resp)
	}
}

func DetailHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req videologic.VideoDetailReq
		if err := c.ShouldBindUri(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		l := videologic.NewDetailVideoLogic(c.Request.Context(), svcCtx)
		resp, err := l.Detail(&req)
		if err != nil {
			response.Fail(c, http.StatusNotFound, err.Error())
			return
		}
		response.Ok(c, resp)
	}
}

func UploadHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		title := c.PostForm("title")
		description := c.PostForm("description")
		tags := c.PostForm("tags")

		if title == "" {
			response.Fail(c, http.StatusBadRequest, "标题不能为空")
			return
		}

		videoFile, err := c.FormFile("video")
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "请上传视频文件")
			return
		}

		// Create video record first to get ID
		// TODO: 接入审核后改回 "pending"，由 admin 审核接口（PUT /admin/videos/:id/status）置为 approved
		video := model.Video{
			Title:       title,
			Description: description,
			Tags:        tags,
			AuthorID:    userID,
			Status:      "pending",
		}
		if err := svcCtx.DB.Create(&video).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "创建视频记录失败")
			return
		}

		// Save uploaded video to temp path
		videoDir := filepath.Join(svcCtx.VideoDir, "videos", strconv.FormatUint(uint64(video.ID), 10))
		os.MkdirAll(videoDir, 0755)

		tmpPath := filepath.Join(videoDir, "upload"+filepath.Ext(videoFile.Filename))
		src, err := videoFile.Open()
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "读取上传文件失败")
			return
		}
		defer src.Close()

		dst, err := os.Create(tmpPath)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存上传文件失败")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存上传文件失败")
			return
		}

		// Run ffmpeg to generate HLS
		playlistPath := filepath.Join(videoDir, "playlist.m3u8")
		segmentPattern := filepath.Join(videoDir, "segment_%05d.ts")
		cmd := exec.Command("ffmpeg",
			"-i", tmpPath,
			"-c:v", "libx264",
			"-c:a", "aac",
			"-hls_time", "10",
			"-hls_list_size", "0",
			"-hls_segment_filename", segmentPattern,
			"-f", "hls",
			playlistPath,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			os.RemoveAll(videoDir)
			response.Fail(c, http.StatusInternalServerError, fmt.Sprintf("视频转码失败: %s", string(output)))
			return
		}

		// Remove temp file
		os.Remove(tmpPath)

		// Save cover image if provided
		coverURL := ""
		if coverFile, err := c.FormFile("cover"); err == nil {
			coverDir := filepath.Join(svcCtx.VideoDir, "covers", strconv.FormatUint(uint64(video.ID), 10))
			os.MkdirAll(coverDir, 0755)
			coverPath := filepath.Join(coverDir, coverFile.Filename)
			src, _ := coverFile.Open()
			if src != nil {
				defer src.Close()
				dst, _ := os.Create(coverPath)
				if dst != nil {
					defer dst.Close()
					io.Copy(dst, src)
					coverURL = fmt.Sprintf("/media/covers/%d/%s", video.ID, coverFile.Filename)
				}
			}
		}

		// Update video record with paths
		relativePlaylist := fmt.Sprintf("/media/videos/%d/playlist.m3u8", video.ID)
		updates := map[string]interface{}{
			"video_url": relativePlaylist,
		}
		if coverURL != "" {
			updates["cover_url"] = coverURL
		}
		svcCtx.DB.Model(&video).Updates(updates)

		video.VideoURL = relativePlaylist
		video.CoverURL = coverURL
		response.Ok(c, video)
	}
}

func LikeHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var video model.Video
		if err := svcCtx.DB.Select("id", "status").First(&video, videoID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}
		if video.Status != "approved" {
			response.Fail(c, http.StatusForbidden, "视频未通过审核")
			return
		}

		var liked bool

		err = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var like model.Like
			err := tx.Where("user_id = ? AND video_id = ?", userID, videoID).First(&like).Error

			if err == nil {
				if err := tx.Unscoped().Delete(&like).Error; err != nil {
					return err
				}
				liked = false
				return tx.Model(&model.Video{}).
					Where("id = ? AND like_count > 0", videoID).
					UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error
			}

			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			if err := tx.Create(&model.Like{
				UserID:  userID,
				VideoID: uint(videoID),
			}).Error; err != nil {
				return err
			}

			liked = true
			return tx.Model(&model.Video{}).
				Where("id = ?", videoID).
				UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
		})

		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "点赞操作失败")
			return
		}

		response.Ok(c, gin.H{
			"liked": liked,
		})
	}
}

func CollectHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var video model.Video
		if err := svcCtx.DB.Select("id", "status").First(&video, videoID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}
		if video.Status != "approved" {
			response.Fail(c, http.StatusForbidden, "视频未通过审核")
			return
		}

		var collected bool

		err = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var collect model.Collect
			err := tx.Where("user_id = ? AND video_id = ?", userID, videoID).First(&collect).Error

			if err == nil {
				if err := tx.Unscoped().Delete(&collect).Error; err != nil {
					return err
				}
				collected = false
				return tx.Model(&model.Video{}).
					Where("id = ? AND collect_count > 0", videoID).
					UpdateColumn("collect_count", gorm.Expr("collect_count - ?", 1)).Error
			}

			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			if err := tx.Create(&model.Collect{
				UserID:  userID,
				VideoID: uint(videoID),
			}).Error; err != nil {
				return err
			}

			collected = true
			return tx.Model(&model.Video{}).
				Where("id = ?", videoID).
				UpdateColumn("collect_count", gorm.Expr("collect_count + ?", 1)).Error
		})

		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "收藏操作失败")
			return
		}

		response.Ok(c, gin.H{
			"collected": collected,
		})
	}
}

func AdminListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req adminVideoListReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 10
		}
		if req.PageSize > 100 {
			req.PageSize = 100
		}

		db := svcCtx.DB.Model(&model.Video{}).Preload("Author")

		if req.Status != "" {
			if !isValidVideoStatus(req.Status) {
				response.Fail(c, http.StatusBadRequest, "无效的视频状态")
				return
			}
			db = db.Where("status = ?", req.Status)
		}

		if req.Keyword != "" {
			keyword := "%" + req.Keyword + "%"
			db = db.Where("title LIKE ? OR description LIKE ?", keyword, keyword)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "视频列表加载失败")
			return
		}

		var videos []model.Video
		if err := db.
			Order("created_at DESC").
			Offset((req.Page - 1) * req.PageSize).
			Limit(req.PageSize).
			Find(&videos).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "视频列表加载失败")
			return
		}

		list := make([]videologic.VideoInfo, 0, len(videos))
		for _, video := range videos {
			list = append(list, videologic.VideoInfo{
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
			})
		}

		response.Ok(c, videologic.PageResult[videologic.VideoInfo]{
			List:     list,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		})
	}
}

func AdminUpdateStatusHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var req updateVideoStatusReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		if !isValidVideoStatus(req.Status) {
			response.Fail(c, http.StatusBadRequest, "无效的视频状态")
			return
		}

		result := svcCtx.DB.Model(&model.Video{}).
			Where("id = ?", videoID).
			Update("status", req.Status)

		if result.Error != nil {
			response.Fail(c, http.StatusInternalServerError, "视频状态更新失败")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}

		response.Ok(c, gin.H{
			"id":     videoID,
			"status": req.Status,
		})
	}
}
