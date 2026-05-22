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
			Status:      "approved",
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
