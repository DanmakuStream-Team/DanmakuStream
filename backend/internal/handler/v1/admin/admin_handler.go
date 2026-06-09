package admin

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/metrics"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
)

type pageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type adminUserItem struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Bio          string `json:"bio"`
	Role         string `json:"role"`
	FollowCount  int64  `json:"followCount"`
	FanCount     int64  `json:"fanCount"`
	VideoCount   int64  `json:"videoCount"`
	DanmakuCount int64  `json:"danmakuCount"`
	CreatedAt    string `json:"createdAt"`
}

type updateRoleReq struct {
	Role string `json:"role" binding:"required"`
}

type saveBannerReq struct {
	Title    string `json:"title" binding:"required"`
	ImageURL string `json:"imageUrl"`
	Link     string `json:"link"`
	Enabled  bool   `json:"enabled"`
	Sort     int    `json:"sort"`
}

type saveAnnouncementReq struct {
	Content   string `json:"content" binding:"required"`
	Enabled   bool   `json:"enabled"`
	StartedAt string `json:"startedAt"`
	EndedAt   string `json:"endedAt"`
}

func InfrastructureHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		usage := diskUsage(svcCtx.VideoDir)
		cpu := cpuUsage()

		var liveViewerCount int64
		_ = svcCtx.DB.Model(&model.LiveRoom{}).
			Where("status = ?", "live").
			Select("COALESCE(SUM(viewer_count), 0)").
			Scan(&liveViewerCount).Error

		var highestConcurrent int64
		_ = svcCtx.DB.Model(&model.LiveRoom{}).
			Select("COALESCE(MAX(viewer_count), 0)").
			Scan(&highestConcurrent).Error

		var liveRoomCount int64
		_ = svcCtx.DB.Model(&model.LiveRoom{}).
			Where("status = ?", "live").
			Count(&liveRoomCount).Error

		today := time.Now().Format("2006-01-02")
		monthPrefix := time.Now().Format("2006-01")
		var todayDownBytes, monthDownBytes uint64
		_ = svcCtx.DB.Model(&model.TrafficStat{}).
			Where("date = ?", today).
			Select("COALESCE(SUM(bytes), 0)").
			Scan(&todayDownBytes).Error
		_ = svcCtx.DB.Model(&model.TrafficStat{}).
			Where("date LIKE ?", monthPrefix+"%").
			Select("COALESCE(SUM(bytes), 0)").
			Scan(&monthDownBytes).Error

		response.Ok(c, gin.H{
			"storage": gin.H{
				"path":         svcCtx.VideoDir,
				"usedBytes":    usage.used,
				"totalBytes":   usage.total,
				"freeBytes":    usage.free,
				"usagePercent": usage.percent,
				"warning":      usage.percent >= 85,
				"critical":     usage.percent >= 95,
			},
			"traffic": gin.H{
				"todayDownBytes": todayDownBytes,
				"monthDownBytes": monthDownBytes,
				"source":         "go-application-middleware",
			},
			"cpu": gin.H{
				"usagePercent": cpu,
				"warning":      cpu >= 75,
				"critical":     cpu >= 90,
				"source":       "/proc/stat",
			},
			"online": gin.H{
				"current":           liveViewerCount + metrics.ActiveVideoConnections(),
				"highestConcurrent": highestConcurrent,
				"liveRoomCount":     liveRoomCount,
				"liveViewerCount":   liveViewerCount,
				"videoConnections":  metrics.ActiveVideoConnections(),
			},
		})
	}
}

func UserListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := getPage(c)
		keyword := strings.TrimSpace(c.Query("keyword"))

		db := svcCtx.DB.Model(&model.User{})
		if keyword != "" {
			like := "%" + keyword + "%"
			db = db.Where("username LIKE ? OR nickname LIKE ? OR bio LIKE ?", like, like, like)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "用户列表加载失败")
			return
		}

		var users []model.User
		if err := db.Order("created_at DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&users).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "用户列表加载失败")
			return
		}

		list := make([]adminUserItem, 0, len(users))
		for _, user := range users {
			var videoCount, danmakuCount int64
			_ = svcCtx.DB.Model(&model.Video{}).Where("author_id = ?", user.ID).Count(&videoCount).Error
			_ = svcCtx.DB.Model(&model.Danmaku{}).Where("user_id = ?", user.ID).Count(&danmakuCount).Error
			list = append(list, adminUserItem{
				ID:           user.ID,
				Username:     user.Username,
				Nickname:     user.Nickname,
				Avatar:       user.Avatar,
				Bio:          user.Bio,
				Role:         user.Role,
				FollowCount:  user.FollowCount,
				FanCount:     user.FanCount,
				VideoCount:   videoCount,
				DanmakuCount: danmakuCount,
				CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		response.Ok(c, pageResult[adminUserItem]{List: list, Total: total, Page: page, PageSize: pageSize})
	}
}

func UpdateUserRoleHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的用户 ID")
			return
		}

		var req updateRoleReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		if req.Role != "user" && req.Role != "moderator" && req.Role != "admin" {
			response.Fail(c, http.StatusBadRequest, "角色不支持")
			return
		}

		if err := svcCtx.DB.Model(&model.User{}).
			Where("id = ?", id).
			Update("role", req.Role).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "角色更新失败")
			return
		}
		response.Ok(c, gin.H{"id": id, "role": req.Role})
	}
}

func BannerListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []model.SiteBanner
		if err := svcCtx.DB.Order("sort ASC, created_at DESC").Find(&list).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "横幅列表加载失败")
			return
		}
		response.Ok(c, list)
	}
}

func CreateBannerHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req saveBannerReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		item := model.SiteBanner{
			Title:    strings.TrimSpace(req.Title),
			ImageURL: strings.TrimSpace(req.ImageURL),
			Link:     strings.TrimSpace(req.Link),
			Enabled:  req.Enabled,
			Sort:     req.Sort,
		}
		if item.Title == "" {
			response.Fail(c, http.StatusBadRequest, "标题不能为空")
			return
		}
		if err := svcCtx.DB.Create(&item).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "横幅创建失败")
			return
		}
		response.Ok(c, item)
	}
}

func UpdateBannerHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的横幅 ID")
			return
		}
		var req saveBannerReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		updates := map[string]any{
			"title":     strings.TrimSpace(req.Title),
			"image_url": strings.TrimSpace(req.ImageURL),
			"link":      strings.TrimSpace(req.Link),
			"enabled":   req.Enabled,
			"sort":      req.Sort,
		}
		if err := svcCtx.DB.Model(&model.SiteBanner{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "横幅更新失败")
			return
		}
		response.Ok(c, gin.H{"id": id})
	}
}

func DeleteBannerHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return deleteByID(svcCtx, &model.SiteBanner{}, "横幅删除失败")
}

func AnnouncementListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []model.SiteAnnouncement
		if err := svcCtx.DB.Order("created_at DESC").Find(&list).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "公告列表加载失败")
			return
		}
		response.Ok(c, list)
	}
}

func CreateAnnouncementHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req saveAnnouncementReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		item, err := buildAnnouncement(req)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		if err := svcCtx.DB.Create(&item).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "公告创建失败")
			return
		}
		response.Ok(c, item)
	}
}

func UpdateAnnouncementHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的公告 ID")
			return
		}
		var req saveAnnouncementReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		item, err := buildAnnouncement(req)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		updates := map[string]any{
			"content":    item.Content,
			"enabled":    item.Enabled,
			"started_at": item.StartedAt,
			"ended_at":   item.EndedAt,
		}
		if err := svcCtx.DB.Model(&model.SiteAnnouncement{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "公告更新失败")
			return
		}
		response.Ok(c, gin.H{"id": id})
	}
}

func DeleteAnnouncementHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return deleteByID(svcCtx, &model.SiteAnnouncement{}, "公告删除失败")
}

func buildAnnouncement(req saveAnnouncementReq) (model.SiteAnnouncement, error) {
	item := model.SiteAnnouncement{
		Content: strings.TrimSpace(req.Content),
		Enabled: req.Enabled,
	}
	if item.Content == "" {
		return item, errText("公告内容不能为空")
	}
	if req.StartedAt != "" {
		t, err := parseTime(req.StartedAt)
		if err != nil {
			return item, errText("开始时间格式错误")
		}
		item.StartedAt = &t
	}
	if req.EndedAt != "" {
		t, err := parseTime(req.EndedAt)
		if err != nil {
			return item, errText("结束时间格式错误")
		}
		item.EndedAt = &t
	}
	return item, nil
}

func deleteByID(svcCtx *svc.ServiceContext, value any, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效 ID")
			return
		}
		if err := svcCtx.DB.Delete(value, id).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, message)
			return
		}
		response.Ok(c, gin.H{"id": id})
	}
}

type diskStat struct {
	used    uint64
	total   uint64
	free    uint64
	percent float64
}

func diskUsage(path string) diskStat {
	var stat syscall.Statfs_t
	target := path
	if _, err := os.Stat(target); err != nil {
		target = filepath.Dir(target)
	}
	if err := syscall.Statfs(target, &stat); err != nil {
		return diskStat{}
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	percent := 0.0
	if total > 0 {
		percent = float64(used) / float64(total) * 100
	}
	return diskStat{used: used, total: total, free: free, percent: percent}
}

func cpuUsage() float64 {
	first, err := readCPUStat()
	if err != nil {
		return 0
	}
	time.Sleep(120 * time.Millisecond)
	second, err := readCPUStat()
	if err != nil {
		return 0
	}

	totalDelta := second.total - first.total
	idleDelta := second.idle - first.idle
	if totalDelta == 0 {
		return 0
	}
	usage := (float64(totalDelta-idleDelta) / float64(totalDelta)) * 100
	if usage < 0 {
		return 0
	}
	if usage > 100 {
		return 100
	}
	return usage
}

type cpuStat struct {
	total uint64
	idle  uint64
}

func readCPUStat() (cpuStat, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return cpuStat{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return cpuStat{}, scanner.Err()
	}

	parts := strings.Fields(scanner.Text())
	if len(parts) < 5 || parts[0] != "cpu" {
		return cpuStat{}, errText("invalid cpu stat")
	}

	var values []uint64
	for _, part := range parts[1:] {
		value, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return cpuStat{}, err
		}
		values = append(values, value)
	}

	var total uint64
	for _, value := range values {
		total += value
	}
	idle := values[3]
	if len(values) > 4 {
		idle += values[4]
	}
	return cpuStat{total: total, idle: idle}, nil
}

func getPage(c *gin.Context) (int, int) {
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

func parseTime(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02 15:04:05", value)
}

type errText string

func (e errText) Error() string {
	return string(e)
}
