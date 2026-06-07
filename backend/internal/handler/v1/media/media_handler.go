package media

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
)

const maxImageSize = 10 * 1024 * 1024

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

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
