package media

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"danmakustream/user-service/internal/handler/response"
	"danmakustream/user-service/internal/middleware"
	"danmakustream/user-service/internal/svc"

	"github.com/gin-gonic/gin"
)

const maxImageSize = 10 * 1024 * 1024
const maxMessageVideoSize = 200 * 1024 * 1024

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

var allowedImageTypes = map[string]bool{
	"common":  true,
	"dynamic": true,
	"live":    true,
	"cover":   true,
}

func UploadImageHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		imageFile, err := c.FormFile("image")
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "请上传图片文件")
			return
		}
		if imageFile.Size <= 0 || imageFile.Size > maxImageSize {
			response.Fail(c, http.StatusBadRequest, "图片大小不能超过 10MB")
			return
		}

		ext := strings.ToLower(filepath.Ext(imageFile.Filename))
		if !allowedImageExts[ext] {
			response.Fail(c, http.StatusBadRequest, "图片格式只支持 jpg、png、webp、gif")
			return
		}

		imageType := strings.TrimSpace(c.PostForm("type"))
		if imageType == "" {
			imageType = "common"
		}
		if !allowedImageTypes[imageType] {
			response.Fail(c, http.StatusBadRequest, "无效的图片用途")
			return
		}

		src, err := imageFile.Open()
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "读取图片失败")
			return
		}
		defer src.Close()

		head := make([]byte, 512)
		n, _ := src.Read(head)
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			response.Fail(c, http.StatusInternalServerError, "读取图片失败")
			return
		}
		contentType := http.DetectContentType(head[:n])
		if !strings.HasPrefix(contentType, "image/") {
			response.Fail(c, http.StatusBadRequest, "上传文件不是图片")
			return
		}

		randomName, err := randomHex(8)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "生成文件名失败")
			return
		}

		dateDir := time.Now().Format("20060102")
		saveDir := filepath.Join(svcCtx.VideoDir, "images", imageType, dateDir)
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			response.Fail(c, http.StatusInternalServerError, "创建图片目录失败")
			return
		}

		fileName := randomName + ext
		savePath := filepath.Join(saveDir, fileName)
		dst, err := os.Create(savePath)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存图片失败")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存图片失败")
			return
		}

		url := "/media/images/" + imageType + "/" + dateDir + "/" + fileName
		response.Ok(c, gin.H{
			"url":         url,
			"type":        imageType,
			"contentType": contentType,
			"size":        imageFile.Size,
			"userId":      userID,
		})
	}
}

// UploadMessageMediaHandler stores private-message attachments without creating a video post.
func UploadMessageMediaHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		fileHeader, err := c.FormFile("file")
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "请选择图片或视频")
			return
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		mediaType := ""
		maxSize := int64(maxImageSize)
		if allowedImageExts[ext] {
			mediaType = "image"
		} else if ext == ".mp4" || ext == ".webm" {
			mediaType = "video"
			maxSize = maxMessageVideoSize
		}
		if mediaType == "" {
			response.Fail(c, http.StatusBadRequest, "私信附件只支持 jpg、png、webp、gif、mp4、webm")
			return
		}
		if fileHeader.Size <= 0 || fileHeader.Size > maxSize {
			limit := "10MB"
			if mediaType == "video" {
				limit = "200MB"
			}
			response.Fail(c, http.StatusBadRequest, "文件大小不能超过 "+limit)
			return
		}

		src, err := fileHeader.Open()
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "读取附件失败")
			return
		}
		defer src.Close()
		head := make([]byte, 512)
		n, _ := src.Read(head)
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			response.Fail(c, http.StatusInternalServerError, "读取附件失败")
			return
		}
		contentType := http.DetectContentType(head[:n])
		if mediaType == "image" && !strings.HasPrefix(contentType, "image/") {
			response.Fail(c, http.StatusBadRequest, "上传文件不是有效图片")
			return
		}
		if mediaType == "video" && !validMessageVideoHeader(ext, head[:n]) {
			response.Fail(c, http.StatusBadRequest, "上传文件不是有效视频")
			return
		}

		randomName, err := randomHex(16)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "生成文件名失败")
			return
		}
		dateDir := time.Now().Format("20060102")
		userDir := strconv.FormatUint(uint64(userID), 10)
		saveDir := filepath.Join(svcCtx.VideoDir, "messages", userDir, dateDir)
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			response.Fail(c, http.StatusInternalServerError, "创建附件目录失败")
			return
		}
		fileName := randomName + ext
		savePath := filepath.Join(saveDir, fileName)
		dst, err := os.Create(savePath)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存附件失败")
			return
		}
		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			_ = os.Remove(savePath)
			response.Fail(c, http.StatusInternalServerError, "保存附件失败")
			return
		}
		if err := dst.Close(); err != nil {
			_ = os.Remove(savePath)
			response.Fail(c, http.StatusInternalServerError, "保存附件失败")
			return
		}

		response.Ok(c, gin.H{
			"url":         "/media/messages/" + userDir + "/" + dateDir + "/" + fileName,
			"mediaType":   mediaType,
			"name":        filepath.Base(fileHeader.Filename),
			"contentType": contentType,
			"size":        fileHeader.Size,
		})
	}
}

func validMessageVideoHeader(ext string, head []byte) bool {
	if ext == ".mp4" {
		return len(head) >= 12 && string(head[4:8]) == "ftyp"
	}
	return ext == ".webm" && len(head) >= 4 && head[0] == 0x1a && head[1] == 0x45 && head[2] == 0xdf && head[3] == 0xa3
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
