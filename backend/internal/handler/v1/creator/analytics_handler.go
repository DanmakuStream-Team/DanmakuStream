package creator

import (
	"net/http"
	"strconv"
	"time"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
)

type analyticsPoint struct {
	Date        string `json:"date"`
	Views       int64  `json:"views"`
	Collects    int64  `json:"collects"`
	GrowthSpeed int64  `json:"growthSpeed"`
	Streams     int64  `json:"streams"`
}

type dailyMetric struct {
	ViewDelta    int64
	CollectDelta int64
}

type topVideoInfo struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	CoverURL     string `json:"coverUrl"`
	Status       string `json:"status"`
	ViewCount    int64  `json:"viewCount"`
	LikeCount    int64  `json:"likeCount"`
	CollectCount int64  `json:"collectCount"`
}

func AnalyticsHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		creatorID := c.GetUint(middleware.CtxKeyUserID)
		if creatorID == 0 {
			response.Fail(c, http.StatusUnauthorized, "请先登录")
			return
		}

		days, err := strconv.Atoi(c.DefaultQuery("days", "30"))
		if err != nil || (days != 7 && days != 30 && days != 90) {
			response.Fail(c, http.StatusBadRequest, "days 仅支持 7、30 或 90")
			return
		}

		var videoID uint64
		if rawVideoID := c.Query("videoId"); rawVideoID != "" {
			videoID, err = strconv.ParseUint(rawVideoID, 10, 64)
			if err != nil || videoID == 0 {
				response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
				return
			}
			var count int64
			if err := svcCtx.DB.Model(&model.Video{}).
				Where("id = ? AND author_id = ?", videoID, creatorID).
				Count(&count).Error; err != nil || count == 0 {
				response.Fail(c, http.StatusNotFound, "作品不存在或不属于当前用户")
				return
			}
		}

		today := beginningOfDay(time.Now())
		start := today.AddDate(0, 0, -(days - 1))
		endDate := today.Format("2006-01-02")

		var creatorStats []model.CreatorDailyStat
		if err := svcCtx.DB.
			Where("creator_id = ? AND date BETWEEN ? AND ?", creatorID, start.Format("2006-01-02"), endDate).
			Order("date ASC").
			Find(&creatorStats).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "创作数据加载失败")
			return
		}

		byDate := make(map[string]dailyMetric, len(creatorStats))
		streamsByDate := make(map[string]int64, len(creatorStats))
		for _, stat := range creatorStats {
			streamsByDate[stat.Date] = stat.StreamCount
			if videoID == 0 {
				byDate[stat.Date] = dailyMetric{ViewDelta: stat.ViewDelta, CollectDelta: stat.CollectDelta}
			}
		}

		if videoID != 0 {
			var videoStats []model.VideoDailyStat
			if err := svcCtx.DB.
				Where("creator_id = ? AND video_id = ? AND date BETWEEN ? AND ?", creatorID, videoID, start.Format("2006-01-02"), endDate).
				Order("date ASC").
				Find(&videoStats).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "作品数据加载失败")
				return
			}
			for _, stat := range videoStats {
				byDate[stat.Date] = dailyMetric{ViewDelta: stat.ViewDelta, CollectDelta: stat.CollectDelta}
			}
		}

		points := make([]analyticsPoint, 0, days)
		var rangeViews, rangeCollects, rangeStreams int64
		for i := 0; i < days; i++ {
			date := start.AddDate(0, 0, i).Format("2006-01-02")
			metric := byDate[date]
			growth := metric.ViewDelta
			if metric.CollectDelta > 0 {
				growth += metric.CollectDelta
			}
			points = append(points, analyticsPoint{
				Date:        date,
				Views:       metric.ViewDelta,
				Collects:    metric.CollectDelta,
				GrowthSpeed: growth,
				Streams:     streamsByDate[date],
			})
			rangeViews += metric.ViewDelta
			rangeCollects += metric.CollectDelta
			rangeStreams += streamsByDate[date]
		}

		var videoTotals struct {
			Views    int64
			Collects int64
		}
		videoTotalsQuery := svcCtx.DB.Model(&model.Video{}).
			Select("COALESCE(SUM(view_count), 0) AS views, COALESCE(SUM(collect_count), 0) AS collects").
			Where("author_id = ?", creatorID)
		if videoID != 0 {
			videoTotalsQuery = videoTotalsQuery.Where("id = ?", videoID)
		}
		if err := videoTotalsQuery.Scan(&videoTotals).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "作品汇总加载失败")
			return
		}

		topVideos := make([]topVideoInfo, 0, 5)
		if err := svcCtx.DB.Model(&model.Video{}).
			Select("id, title, cover_url, status, view_count, like_count, collect_count").
			Where("author_id = ?", creatorID).
			Order("view_count DESC, collect_count DESC, created_at DESC").
			Limit(5).
			Scan(&topVideos).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "作品排行加载失败")
			return
		}

		var streamTotal struct {
			Total int64
		}
		if err := svcCtx.DB.Model(&model.CreatorDailyStat{}).
			Where("creator_id = ?", creatorID).
			Select("COALESCE(SUM(stream_count), 0) AS total").
			Scan(&streamTotal).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "推流数据加载失败")
			return
		}

		response.Ok(c, gin.H{
			"days":            days,
			"selectedVideoId": videoID,
			"summary": gin.H{
				"totalViews":        videoTotals.Views,
				"totalCollects":     videoTotals.Collects,
				"totalStreams":      streamTotal.Total,
				"rangeViews":        rangeViews,
				"rangeCollects":     rangeCollects,
				"rangeStreams":      rangeStreams,
				"averageDailyViews": float64(rangeViews) / float64(days),
			},
			"points":    points,
			"topVideos": topVideos,
		})
	}
}

func beginningOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}
