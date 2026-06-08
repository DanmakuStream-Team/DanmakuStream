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
	"strings"

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

type updateVideoReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
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
		if err := os.MkdirAll(videoDir, 0755); err != nil {
			response.Fail(c, http.StatusInternalServerError, "创建视频目录失败")
			return
		}

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
		src.Close()
		dst.Close()

		// Save cover image if provided
		coverURL := ""
		if coverFile, err := c.FormFile("cover"); err == nil {
			coverDir := filepath.Join(svcCtx.VideoDir, "covers", strconv.FormatUint(uint64(video.ID), 10))
			if err := os.MkdirAll(coverDir, 0755); err != nil {
				response.Fail(c, http.StatusInternalServerError, "创建封面目录失败")
				return
			}
			coverPath := filepath.Join(coverDir, coverFile.Filename)
			src, _ := coverFile.Open()
			if src != nil {
				dst, _ := os.Create(coverPath)
				if dst != nil {
					io.Copy(dst, src)
					dst.Close()
					coverURL = fmt.Sprintf("/media/covers/%d/%s", video.ID, coverFile.Filename)
				}
				src.Close()
			}
		}

		if coverURL != "" {
			video.CoverURL = coverURL
			svcCtx.DB.Model(&video).Update("cover_url", coverURL)
		}

		go transcodeVideoAsync(svcCtx, video.ID, videoDir, tmpPath, coverURL)

		response.Ok(c, video)
	}
}

func DownloadHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)

		videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var video model.Video
		if err := svcCtx.DB.First(&video, videoID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}

		if video.Status != "approved" && video.AuthorID != userID && role != "admin" {
			response.Fail(c, http.StatusForbidden, "无权下载该视频")
			return
		}
		if video.VideoURL == "" {
			response.Fail(c, http.StatusBadRequest, "视频尚未转码完成，暂时无法下载")
			return
		}

		sourcePath, err := mediaPathToLocalPath(svcCtx.VideoDir, video.VideoURL)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "视频地址无效")
			return
		}
		if _, err := os.Stat(sourcePath); err != nil {
			response.Fail(c, http.StatusNotFound, "视频文件不存在或尚未生成")
			return
		}

		fileName := safeDownloadName(video.Title) + ".mp4"
		if strings.HasSuffix(strings.ToLower(sourcePath), ".m3u8") {
			tmpFile, err := os.CreateTemp("", "danmakustream-download-*.mp4")
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, "创建下载文件失败")
				return
			}
			tmpPath := tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpPath)

			cmd := exec.Command("ffmpeg",
				"-y",
				"-allowed_extensions", "ALL",
				"-i", sourcePath,
				"-c", "copy",
				"-bsf:a", "aac_adtstoasc",
				tmpPath,
			)
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Printf("video %d download remux failed: %s\n", video.ID, string(output))
				response.Fail(c, http.StatusInternalServerError, "生成下载文件失败")
				return
			}

			c.FileAttachment(tmpPath, fileName)
			return
		}

		c.FileAttachment(sourcePath, fileName)
	}
}

func mediaPathToLocalPath(videoDir string, url string) (string, error) {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return "", errors.New("remote media is not supported")
	}
	cleanURL := strings.TrimPrefix(url, "/")
	const mediaPrefix = "media/"
	if !strings.HasPrefix(cleanURL, mediaPrefix) {
		return "", errors.New("invalid media path")
	}
	relative := strings.TrimPrefix(cleanURL, mediaPrefix)
	return filepath.Join(videoDir, filepath.FromSlash(relative)), nil
}

func safeDownloadName(title string) string {
	name := strings.TrimSpace(title)
	if name == "" {
		name = "danmaku-video"
	}
	replacer := strings.NewReplacer(
		"\\", "_",
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}

func transcodeVideoAsync(svcCtx *svc.ServiceContext, videoID uint, videoDir, tmpPath, coverURL string) {
	playlistPath := filepath.Join(videoDir, "playlist.m3u8")
	segmentPattern := filepath.Join(videoDir, "segment_%05d.ts")
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", tmpPath,
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-c:a", "aac",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", segmentPattern,
		"-f", "hls",
		playlistPath,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("video %d transcode failed: %s\n", videoID, string(output))
		return
	}

	thumbnailName := "thumbnail.jpg"
	thumbnailPath := filepath.Join(videoDir, thumbnailName)
	thumbCmd := exec.Command("ffmpeg",
		"-y",
		"-i", tmpPath,
		"-frames:v", "1",
		"-q:v", "2",
		thumbnailPath,
	)
	thumbCmd.CombinedOutput()
	os.Remove(tmpPath)

	relativePlaylist := fmt.Sprintf("/media/videos/%d/playlist.m3u8", videoID)
	updates := map[string]any{
		"video_url": relativePlaylist,
	}
	if coverURL == "" {
		if _, err := os.Stat(thumbnailPath); err == nil {
			updates["cover_url"] = fmt.Sprintf("/media/videos/%d/%s", videoID, thumbnailName)
		}
	}
	svcCtx.DB.Model(&model.Video{}).Where("id = ?", videoID).Updates(updates)
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

		if req.Status == "approved" {
			var video model.Video
			if err := svcCtx.DB.Select("id", "video_url").First(&video, videoID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					response.Fail(c, http.StatusNotFound, "视频不存在")
					return
				}
				response.Fail(c, http.StatusInternalServerError, "视频加载失败")
				return
			}
			if video.VideoURL == "" {
				response.Fail(c, http.StatusForbidden, "视频仍在转码，暂时不能审核通过")
				return
			}
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

func UpdateHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)

		videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var req updateVideoReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		var video model.Video
		if err := svcCtx.DB.First(&video, videoID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}

		if video.AuthorID != userID && role != "admin" {
			response.Fail(c, http.StatusForbidden, "无权修改该视频")
			return
		}

		updates := map[string]interface{}{}

		if req.Title != "" {
			updates["title"] = req.Title
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		if req.Tags != "" {
			updates["tags"] = req.Tags
		}

		if len(updates) == 0 {
			response.Fail(c, http.StatusBadRequest, "没有需要更新的内容")
			return
		}

		if err := svcCtx.DB.Model(&video).Updates(updates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "视频信息更新失败")
			return
		}

		response.Ok(c, gin.H{
			"id": video.ID,
		})
	}
}

func UpdateCoverHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)

		videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var video model.Video
		if err := svcCtx.DB.First(&video, videoID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}

		if video.AuthorID != userID && role != "admin" {
			response.Fail(c, http.StatusForbidden, "无权修改该视频")
			return
		}

		coverFile, err := c.FormFile("cover")
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "请上传封面图片")
			return
		}

		fileExt := filepath.Ext(coverFile.Filename)
		if fileExt == "" {
			response.Fail(c, http.StatusBadRequest, "封面文件格式无效")
			return
		}

		coverDir := filepath.Join(svcCtx.VideoDir, "covers", strconv.FormatUint(videoID, 10))
		if err := os.MkdirAll(coverDir, 0755); err != nil {
			response.Fail(c, http.StatusInternalServerError, "创建封面目录失败")
			return
		}

		fileName := "cover" + fileExt
		coverPath := filepath.Join(coverDir, fileName)

		src, err := coverFile.Open()
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "读取封面文件失败")
			return
		}
		defer src.Close()

		dst, err := os.Create(coverPath)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存封面文件失败")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存封面文件失败")
			return
		}

		coverURL := fmt.Sprintf("/media/covers/%d/%s", videoID, fileName)

		if err := svcCtx.DB.Model(&video).
			Update("cover_url", coverURL).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "封面更新失败")
			return
		}

		response.Ok(c, gin.H{
			"coverUrl": coverURL,
		})
	}
}

func DeleteHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)

		videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var video model.Video
		if err := svcCtx.DB.First(&video, videoID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}

		if video.AuthorID != userID && role != "admin" {
			response.Fail(c, http.StatusForbidden, "无权删除该视频")
			return
		}

		if err := svcCtx.DB.Delete(&video).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "视频删除失败")
			return
		}

		videoDir := filepath.Join(svcCtx.VideoDir, "videos", strconv.FormatUint(videoID, 10))
		coverDir := filepath.Join(svcCtx.VideoDir, "covers", strconv.FormatUint(videoID, 10))

		_ = os.RemoveAll(videoDir)
		_ = os.RemoveAll(coverDir)

		response.Ok(c, gin.H{
			"id": video.ID,
		})
	}
}
