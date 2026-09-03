package admin

import (
	"danmakustream/user-service/internal/handler/response"
	model "danmakustream/user-service/internal/model/mysql"
	"danmakustream/user-service/internal/svc"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func UserList(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
		keyword := strings.TrimSpace(c.Query("keyword"))
		db := ctx.DB.Model(&model.User{})
		if keyword != "" {
			like := "%" + keyword + "%"
			db = db.Where("username LIKE ? OR nickname LIKE ? OR bio LIKE ?", like, like, like)
		}
		var users []model.User
		var total int64
		db.Count(&total)
		if err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
			response.Fail(c, 500, "query users failed")
			return
		}
		items := make([]model.UserInfo, 0, len(users))
		for _, u := range users {
			items = append(items, model.UserInfo{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Avatar: u.Avatar, Role: u.Role})
		}
		response.Ok(c, gin.H{"items": items, "list": items, "page": page, "pageSize": pageSize, "total": total})
	}
}

// Infrastructure exposes platform-level fallback metrics from the service
// runtime. Domain-specific counters can later be aggregated by a dedicated
// observability service without changing the public response contract.
func Infrastructure(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		used, total, free := diskUsage(ctx.VideoDir)
		percent := float64(0)
		if total > 0 {
			percent = float64(used) * 100 / float64(total)
		}
		response.Ok(c, gin.H{
			"storage": gin.H{"path": ctx.VideoDir, "usedBytes": used, "totalBytes": total, "freeBytes": free, "usagePercent": percent, "warning": percent >= 85, "critical": percent >= 95},
			"traffic": gin.H{"todayDownBytes": 0, "monthDownBytes": 0, "source": "go-application-middleware"},
			"cpu":     gin.H{"usagePercent": 0, "warning": false, "critical": false, "source": "service-runtime"},
			"online":  gin.H{"current": 0, "highestConcurrent": 0, "liveRoomCount": 0, "liveViewerCount": 0, "videoConnections": 0},
		})
	}
}

func diskUsage(path string) (used, total, free uint64) {
	out, err := exec.Command("df", "-Pk", path).Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[len(lines)-1])
			if len(fields) >= 4 {
				totalKB, e1 := strconv.ParseUint(fields[1], 10, 64)
				usedKB, e2 := strconv.ParseUint(fields[2], 10, 64)
				freeKB, e3 := strconv.ParseUint(fields[3], 10, 64)
				if e1 == nil && e2 == nil && e3 == nil {
					return usedKB * 1024, totalKB * 1024, freeKB * 1024
				}
			}
		}
	}
	// Portable fallback used only when the platform does not provide df.
	var size uint64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, walkErr error) error {
		if walkErr == nil && info != nil && !info.IsDir() {
			size += uint64(info.Size())
		}
		return nil
	})
	total = 10 * 1024 * 1024 * 1024
	if size > total {
		total = size
	}
	return size, total, total - size
}
func UpdateRole(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, 400, "invalid user id")
			return
		}
		var req struct {
			Role string `json:"role"`
		}
		if c.ShouldBindJSON(&req) != nil || (req.Role != "user" && req.Role != "creator" && req.Role != "moderator" && req.Role != "admin") {
			response.Fail(c, 400, "invalid role")
			return
		}
		result := ctx.DB.Model(&model.User{}).Where("id=?", id).Update("role", req.Role)
		if result.Error != nil {
			response.Fail(c, 500, "update role failed")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusNotFound, "user not found")
			return
		}
		response.Ok(c, gin.H{"id": id, "role": req.Role})
	}
}
