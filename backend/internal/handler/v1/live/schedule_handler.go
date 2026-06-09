package live

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type createScheduleReq struct {
	Title       string `json:"title" binding:"required"`
	CoverURL    string `json:"coverUrl"`
	ScheduledAt string `json:"scheduledAt" binding:"required"`
}

type liveScheduleInfo struct {
	ID            uint            `json:"id"`
	Title         string          `json:"title"`
	CoverURL      string          `json:"coverUrl"`
	ScheduledAt   string          `json:"scheduledAt"`
	Status        string          `json:"status"`
	ReminderCount int64           `json:"reminderCount"`
	Reserved      bool            `json:"reserved"`
	OwnerID       uint            `json:"ownerId"`
	Owner         *model.UserInfo `json:"owner,omitempty"`
	CreatedAt     string          `json:"createdAt"`
}

func StartScheduleWorker(svcCtx *svc.ServiceContext) {
	go func() {
		processDueLiveSchedules(svcCtx)

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			processDueLiveSchedules(svcCtx)
		}
	}()
}

func ScheduleListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := getPage(c)

		db := svcCtx.DB.Model(&model.LiveSchedule{}).Preload("Owner")
		if status := c.Query("status"); status != "" {
			if !isValidScheduleStatus(status) {
				response.Fail(c, http.StatusBadRequest, "无效的预约状态")
				return
			}
			db = db.Where("status = ?", status)
		} else {
			db = db.Where("status = ?", "pending")
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播预约列表加载失败")
			return
		}

		var schedules []model.LiveSchedule
		if err := db.Order("scheduled_at ASC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&schedules).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播预约列表加载失败")
			return
		}

		reservedMap := map[uint]bool{}
		currentUserID := getOptionalUserID(c, svcCtx)
		if currentUserID != 0 && len(schedules) > 0 {
			ids := make([]uint, 0, len(schedules))
			for _, schedule := range schedules {
				ids = append(ids, schedule.ID)
			}
			var reservations []model.LiveReservation
			if err := svcCtx.DB.
				Where("user_id = ? AND schedule_id IN ?", currentUserID, ids).
				Find(&reservations).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "预约状态加载失败")
				return
			}
			for _, reservation := range reservations {
				reservedMap[reservation.ScheduleID] = true
			}
		}

		list := make([]liveScheduleInfo, 0, len(schedules))
		for _, schedule := range schedules {
			list = append(list, toLiveScheduleInfo(schedule, reservedMap[schedule.ID]))
		}

		response.Ok(c, gin.H{
			"list":     list,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

func CreateScheduleHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		var req createScheduleReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		title := strings.TrimSpace(req.Title)
		if title == "" {
			response.Fail(c, http.StatusBadRequest, "预约标题不能为空")
			return
		}

		scheduledAt, err := parseScheduleTime(req.ScheduledAt)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "预约时间格式错误")
			return
		}
		if scheduledAt.Before(time.Now()) {
			response.Fail(c, http.StatusBadRequest, "预约时间不能早于当前时间")
			return
		}

		schedule := model.LiveSchedule{
			Title:       title,
			CoverURL:    strings.TrimSpace(req.CoverURL),
			ScheduledAt: scheduledAt,
			Status:      "pending",
			OwnerID:     userID,
		}

		err = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&schedule).Error; err != nil {
				return err
			}
			return notifyLiveFollowers(tx, userID, "live_schedule", "你关注的主播创建了直播预约", title, "/live-schedules")
		})
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播预约创建失败")
			return
		}

		if err := svcCtx.DB.Preload("Owner").First(&schedule, schedule.ID).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播预约加载失败")
			return
		}

		response.Ok(c, toLiveScheduleInfo(schedule, false))
	}
}

func CancelScheduleHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的预约 ID")
			return
		}

		var schedule model.LiveSchedule
		if err := svcCtx.DB.First(&schedule, id).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "直播预约不存在")
			return
		}

		if schedule.OwnerID != userID && role != "admin" {
			response.Fail(c, http.StatusForbidden, "无权取消该直播预约")
			return
		}

		if err := svcCtx.DB.Model(&schedule).Update("status", "canceled").Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播预约取消失败")
			return
		}

		response.Ok(c, gin.H{"id": schedule.ID, "status": "canceled"})
	}
}

func ReserveScheduleHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的预约 ID")
			return
		}

		var reserved bool
		var reminderCount int64
		err = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var schedule model.LiveSchedule
			if err := tx.First(&schedule, id).Error; err != nil {
				return err
			}
			if schedule.Status != "pending" {
				return errors.New("该直播预约不可预约")
			}
			if schedule.OwnerID == userID {
				return errors.New("不能预约自己的直播")
			}

			var reservation model.LiveReservation
			err := tx.Where("schedule_id = ? AND user_id = ?", id, userID).First(&reservation).Error
			if err == nil {
				if err := tx.Unscoped().Delete(&reservation).Error; err != nil {
					return err
				}
				reserved = false
				if err := tx.Model(&model.LiveSchedule{}).
					Where("id = ? AND reminder_count > 0", id).
					UpdateColumn("reminder_count", gorm.Expr("reminder_count - ?", 1)).Error; err != nil {
					return err
				}
			} else {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if err := tx.Create(&model.LiveReservation{
					ScheduleID: uint(id),
					UserID:     userID,
				}).Error; err != nil {
					return err
				}
				reserved = true
				if err := tx.Model(&model.LiveSchedule{}).
					Where("id = ?", id).
					UpdateColumn("reminder_count", gorm.Expr("reminder_count + ?", 1)).Error; err != nil {
					return err
				}
			}

			return tx.Model(&model.LiveSchedule{}).
				Where("id = ?", id).
				Select("reminder_count").
				Scan(&reminderCount).Error
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, "直播预约不存在")
				return
			}
			response.Fail(c, http.StatusBadRequest, err.Error())
			return
		}

		response.Ok(c, gin.H{
			"reserved":      reserved,
			"reminderCount": reminderCount,
		})
	}
}

func processDueLiveSchedules(svcCtx *svc.ServiceContext) {
	var schedules []model.LiveSchedule
	if err := svcCtx.DB.
		Where("status = ? AND scheduled_at <= ?", "pending", time.Now()).
		Order("scheduled_at ASC").
		Limit(20).
		Find(&schedules).Error; err != nil {
		fmt.Println("live schedule worker load failed:", err)
		return
	}

	for _, schedule := range schedules {
		if err := startScheduledLive(svcCtx, schedule); err != nil {
			fmt.Println("live schedule worker start failed:", err)
		}
	}
}

func startScheduledLive(svcCtx *svc.ServiceContext, schedule model.LiveSchedule) error {
	return svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.LiveSchedule{}).
			Where("id = ? AND status = ?", schedule.ID, "pending").
			Update("status", "live")
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		now := time.Now()
		streamKey, err := generateStreamKey()
		if err != nil {
			return err
		}

		var room model.LiveRoom
		err = tx.Where("owner_id = ?", schedule.OwnerID).First(&room).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			room = model.LiveRoom{
				Title:       schedule.Title,
				CoverURL:    schedule.CoverURL,
				StreamKey:   streamKey,
				Status:      "live",
				ViewerCount: 0,
				OwnerID:     schedule.OwnerID,
				StartedAt:   &now,
				EndedAt:     nil,
			}
			if err := tx.Create(&room).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			if err := tx.Model(&room).Updates(map[string]any{
				"title":        schedule.Title,
				"cover_url":    schedule.CoverURL,
				"stream_key":   streamKey,
				"status":       "live",
				"viewer_count": 0,
				"started_at":   &now,
				"ended_at":     nil,
			}).Error; err != nil {
				return err
			}
		}

		return notifyScheduleReservationUsers(tx, schedule.ID, schedule.OwnerID, schedule.Title, "/live")
	})
}

func notifyScheduleReservationUsers(tx *gorm.DB, scheduleID, ownerID uint, title, link string) error {
	var reservations []model.LiveReservation
	if err := tx.Where("schedule_id = ?", scheduleID).Find(&reservations).Error; err != nil {
		return err
	}
	if len(reservations) == 0 {
		return nil
	}

	actor := ownerID
	notifications := make([]model.Notification, 0, len(reservations))
	for _, reservation := range reservations {
		if reservation.UserID == ownerID {
			continue
		}
		notifications = append(notifications, model.Notification{
			UserID:  reservation.UserID,
			ActorID: &actor,
			Type:    "live_start",
			Title:   "你预约的直播已开播",
			Content: title,
			Link:    link,
		})
	}
	if len(notifications) == 0 {
		return nil
	}
	return tx.Create(&notifications).Error
}

func notifyScheduleReservations(tx *gorm.DB, ownerID uint, title string) error {
	var schedules []model.LiveSchedule
	if err := tx.Where("owner_id = ? AND status = ?", ownerID, "pending").Find(&schedules).Error; err != nil {
		return err
	}
	if len(schedules) == 0 {
		return nil
	}

	actor := ownerID
	for _, schedule := range schedules {
		var reservations []model.LiveReservation
		if err := tx.Where("schedule_id = ?", schedule.ID).Find(&reservations).Error; err != nil {
			return err
		}
		if len(reservations) == 0 {
			continue
		}

		notifications := make([]model.Notification, 0, len(reservations))
		for _, reservation := range reservations {
			notifications = append(notifications, model.Notification{
				UserID:  reservation.UserID,
				ActorID: &actor,
				Type:    "live_start",
				Title:   "你预约的直播已开播",
				Content: title,
				Link:    "/live",
			})
		}
		if err := tx.Create(&notifications).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.LiveSchedule{}).Where("id = ?", schedule.ID).Update("status", "live").Error; err != nil {
			return err
		}
	}
	return nil
}

func notifyLiveFollowers(tx *gorm.DB, actorID uint, noticeType, title, content, link string) error {
	var follows []model.Follow
	if err := tx.Where("followee_id = ?", actorID).Find(&follows).Error; err != nil {
		return err
	}
	if len(follows) == 0 {
		return nil
	}

	actor := actorID
	notifications := make([]model.Notification, 0, len(follows))
	for _, follow := range follows {
		if follow.FollowerID == actorID {
			continue
		}
		notifications = append(notifications, model.Notification{
			UserID:  follow.FollowerID,
			ActorID: &actor,
			Type:    noticeType,
			Title:   title,
			Content: content,
			Link:    link,
		})
	}
	if len(notifications) == 0 {
		return nil
	}
	return tx.Create(&notifications).Error
}

func toLiveScheduleInfo(schedule model.LiveSchedule, reserved bool) liveScheduleInfo {
	return liveScheduleInfo{
		ID:            schedule.ID,
		Title:         schedule.Title,
		CoverURL:      schedule.CoverURL,
		ScheduledAt:   schedule.ScheduledAt.Format("2006-01-02 15:04:05"),
		Status:        schedule.Status,
		ReminderCount: schedule.ReminderCount,
		Reserved:      reserved,
		OwnerID:       schedule.OwnerID,
		Owner: &model.UserInfo{
			ID:       schedule.Owner.ID,
			Username: schedule.Owner.Username,
			Nickname: schedule.Owner.Nickname,
			Avatar:   schedule.Owner.Avatar,
			Role:     schedule.Owner.Role,
		},
		CreatedAt: schedule.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func parseScheduleTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
}

func isValidScheduleStatus(status string) bool {
	return status == "pending" || status == "canceled" || status == "live"
}

func getOptionalUserID(c *gin.Context, svcCtx *svc.ServiceContext) uint {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(svcCtx.Config.Auth.AccessSecret), nil
	})
	if err != nil || !token.Valid {
		return 0
	}

	return claims.UserID
}
