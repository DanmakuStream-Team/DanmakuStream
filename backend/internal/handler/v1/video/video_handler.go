package video

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	videologic "danmakustream/backend/internal/logic/video"
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
