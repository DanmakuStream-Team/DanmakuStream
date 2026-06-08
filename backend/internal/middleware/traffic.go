package middleware

import (
	"strings"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type trafficWriter struct {
	gin.ResponseWriter
	bytes int
}

func (w *trafficWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.bytes += n
	return n, err
}

func (w *trafficWriter) WriteString(data string) (int, error) {
	n, err := w.ResponseWriter.WriteString(data)
	w.bytes += n
	return n, err
}

func TrafficMiddleware(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		writer := &trafficWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()

		if writer.bytes <= 0 {
			return
		}

		category := trafficCategory(c.Request.URL.Path)
		if category == "" {
			return
		}

		date := time.Now().Format("2006-01-02")
		_ = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var stat model.TrafficStat
			err := tx.Where("date = ? AND category = ?", date, category).First(&stat).Error
			if err == nil {
				return tx.Model(&stat).UpdateColumn("bytes", gorm.Expr("bytes + ?", writer.bytes)).Error
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			return tx.Create(&model.TrafficStat{
				Date:     date,
				Category: category,
				Bytes:    uint64(writer.bytes),
			}).Error
		})
	}
}

func trafficCategory(path string) string {
	switch {
	case strings.HasPrefix(path, "/media/videos"):
		return "video"
	case strings.HasPrefix(path, "/media/covers"), strings.HasPrefix(path, "/media/avatars"), strings.HasPrefix(path, "/media/images"):
		return "media"
	case strings.HasPrefix(path, "/api/"):
		return "api"
	default:
		return ""
	}
}
