package live

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"danmakustream/backend/internal/handler/response"
	danmakulogic "danmakustream/backend/internal/logic/danmaku"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type giftDefinition struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

var giftCatalog = map[string]giftDefinition{
	"flower": {Key: "flower", Name: "鲜花", Value: 10},
	"star":   {Key: "star", Name: "星光", Value: 50},
	"rocket": {Key: "rocket", Name: "火箭", Value: 200},
}

type sendGiftReq struct {
	GiftKey string `json:"giftKey" binding:"required"`
	Count   int    `json:"count" binding:"required"`
	Message string `json:"message" binding:"max=200"`
}

type supportRankItem struct {
	UserID    uint            `json:"userId"`
	User      *model.UserInfo `json:"user"`
	Value     int64           `json:"value"`
	GiftCount int64           `json:"giftCount"`
}

type heatRankItem struct {
	Room liveRoomInfo `json:"room"`
	Heat int64        `json:"heat"`
}

// MonitorHandler returns recent room activity for the broadcaster monitor.
func MonitorHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID, err := parseRoomID(c)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "无效的直播间 ID")
			return
		}
		userID := c.GetUint(middleware.CtxKeyUserID)
		var room model.LiveRoom
		if err := svcCtx.DB.Where("id = ? AND owner_id = ? AND status = ?", roomID, userID, "live").First(&room).Error; err != nil {
			response.Fail(c, http.StatusForbidden, "无权查看这个直播间的监看数据")
			return
		}

		danmakuQuery := svcCtx.DB.Where("video_id = ? AND scene = ? AND blocked = ?", roomID, "live", false)
		if room.StartedAt != nil {
			danmakuQuery = danmakuQuery.Where("created_at >= ?", *room.StartedAt)
		}
		var danmakus []model.Danmaku
		if err := danmakuQuery.Order("created_at DESC").Limit(100).Find(&danmakus).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "监看消息加载失败")
			return
		}

		superChats, err := loadRecentSuperChats(svcCtx, room, 50)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "SC 记录加载失败")
			return
		}

		userIDs := make([]uint, 0, len(danmakus))
		for _, item := range danmakus {
			userIDs = append(userIDs, item.UserID)
		}
		var users []model.User
		if len(userIDs) > 0 {
			_ = svcCtx.DB.Where("id IN ?", userIDs).Find(&users).Error
		}
		userMap := make(map[uint]*model.UserInfo, len(users))
		for _, user := range users {
			userMap[user.ID] = &model.UserInfo{ID: user.ID, Username: user.Username, Nickname: user.Nickname, Avatar: user.Avatar, Role: user.Role}
		}

		messages := make([]gin.H, 0, len(danmakus))
		for i := len(danmakus) - 1; i >= 0; i-- {
			item := danmakus[i]
			messages = append(messages, gin.H{
				"id": item.ID, "videoId": item.VideoID, "userId": item.UserID,
				"content": item.Content, "time": item.Time, "color": item.Color,
				"fontSize": item.FontSize, "type": item.Type, "author": userMap[item.UserID],
				"createdAt": item.CreatedAt.Format(time.RFC3339),
			})
		}
		response.Ok(c, gin.H{"messages": messages, "superChats": superChats})
	}
}

func InteractionHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		room, ok := loadLiveRoom(c, svcCtx)
		if !ok {
			return
		}
		rank, err := loadSupportRank(svcCtx, room.ID, 10)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播互动信息加载失败")
			return
		}
		catalog := []giftDefinition{giftCatalog["flower"], giftCatalog["star"], giftCatalog["rocket"]}
		superChats, _ := loadRecentSuperChats(svcCtx, room, 20)
		response.Ok(c, gin.H{
			"likeCount":   room.LikeCount,
			"giftValue":   room.GiftValue,
			"heat":        roomHeat(room),
			"gifts":       catalog,
			"supportRank": rank,
			"superChats":  superChats,
		})
	}
}

func HeatRankingHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rooms []model.LiveRoom
		if err := svcCtx.DB.Preload("Owner").Where("status = ?", "live").Find(&rooms).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播热度排行加载失败")
			return
		}
		items := make([]heatRankItem, 0, len(rooms))
		for _, room := range rooms {
			items = append(items, heatRankItem{
				Room: toLiveRoomInfo(room, false, svcCtx.Config.Live.RTMPHost, svcCtx.Config.Live.HTTPHost),
				Heat: roomHeat(room),
			})
		}
		for i := 0; i < len(items); i++ {
			for j := i + 1; j < len(items); j++ {
				if items[j].Heat > items[i].Heat {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
		if len(items) > 10 {
			items = items[:10]
		}
		response.Ok(c, items)
	}
}

func LikeHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID, err := parseRoomID(c)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "无效的直播间 ID")
			return
		}
		userID := c.GetUint(middleware.CtxKeyUserID)
		var liked bool
		var room model.LiveRoom
		err = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", roomID, "live").First(&room).Error; err != nil {
				return err
			}
			var record model.LiveLike
			err := tx.Unscoped().Where("room_id = ? AND user_id = ?", roomID, userID).First(&record).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				liked = true
				if err := tx.Create(&model.LiveLike{RoomID: roomID, UserID: userID}).Error; err != nil {
					return err
				}
			case err != nil:
				return err
			case record.DeletedAt.Valid:
				liked = true
				if err := tx.Unscoped().Model(&record).Updates(map[string]any{"deleted_at": nil}).Error; err != nil {
					return err
				}
			default:
				liked = false
				if err := tx.Delete(&record).Error; err != nil {
					return err
				}
			}
			var count int64
			if err := tx.Model(&model.LiveLike{}).Where("room_id = ?", roomID).Count(&count).Error; err != nil {
				return err
			}
			room.LikeCount = count
			return tx.Model(&room).Update("like_count", count).Error
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "直播间不存在或已结束")
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "点赞操作失败")
			return
		}
		payload := gin.H{"userId": userID, "liked": liked, "likeCount": room.LikeCount, "heat": roomHeat(room)}
		danmakulogic.GetHub(svcCtx).BroadcastEvent(roomID, "live_like", payload)
		response.Ok(c, payload)
	}
}

func LikeStatusHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID, err := parseRoomID(c)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "无效的直播间 ID")
			return
		}
		userID := c.GetUint(middleware.CtxKeyUserID)
		var count int64
		if err := svcCtx.DB.Model(&model.LiveLike{}).Where("room_id = ? AND user_id = ?", roomID, userID).Count(&count).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "点赞状态加载失败")
			return
		}
		response.Ok(c, gin.H{"liked": count > 0})
	}
}

func CurrentDanmakuHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		room, ok := loadLiveRoom(c, svcCtx)
		if !ok {
			return
		}
		query := svcCtx.DB.Where("video_id = ? AND scene = ? AND blocked = ?", room.ID, "live", false)
		if room.StartedAt != nil {
			query = query.Where("created_at >= ?", *room.StartedAt)
		}
		var items []model.Danmaku
		if err := query.Order("created_at ASC").Limit(300).Find(&items).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播弹幕加载失败")
			return
		}
		result := make([]gin.H, 0, len(items))
		for _, item := range items {
			var user model.User
			svcCtx.DB.First(&user, item.UserID)
			result = append(result, gin.H{
				"id": item.ID, "videoId": item.VideoID, "userId": item.UserID,
				"content": item.Content, "time": item.Time, "color": item.Color,
				"fontSize": item.FontSize, "type": item.Type,
				"author": &model.UserInfo{ID: user.ID, Username: user.Username, Nickname: user.Nickname, Avatar: user.Avatar, Role: user.Role},
			})
		}
		response.Ok(c, result)
	}
}

func GiftHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID, err := parseRoomID(c)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "无效的直播间 ID")
			return
		}
		var req sendGiftReq
		if err := c.ShouldBindJSON(&req); err != nil || req.Count < 1 || req.Count > 99 {
			response.Fail(c, http.StatusBadRequest, "礼物参数错误")
			return
		}
		gift, exists := giftCatalog[req.GiftKey]
		if !exists {
			response.Fail(c, http.StatusBadRequest, "礼物不存在")
			return
		}
		userID := c.GetUint(middleware.CtxKeyUserID)
		value := gift.Value * int64(req.Count)
		message := strings.TrimSpace(req.Message)
		displaySeconds := 0
		if message != "" {
			displaySeconds = superChatDisplaySeconds(value)
		}
		var room model.LiveRoom
		var record model.LiveGift
		err = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", roomID, "live").First(&room).Error; err != nil {
				return err
			}
			record = model.LiveGift{
				RoomID: roomID, UserID: userID, GiftKey: gift.Key, Name: gift.Name,
				Count: req.Count, Value: value, Message: message, DisplaySeconds: displaySeconds,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			room.GiftValue += value
			return tx.Model(&room).Update("gift_value", room.GiftValue).Error
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "直播间不存在或已结束")
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "礼物发送失败")
			return
		}
		var user model.User
		svcCtx.DB.First(&user, userID)
		userInfo := &model.UserInfo{ID: user.ID, Username: user.Username, Nickname: user.Nickname, Avatar: user.Avatar, Role: user.Role}
		rank, _ := loadSupportRank(svcCtx, roomID, 10)
		payload := gin.H{
			"id": record.ID, "createdAt": record.CreatedAt.Format(time.RFC3339),
			"user": userInfo, "gift": gift, "count": req.Count, "value": value,
			"message": message, "displaySeconds": displaySeconds,
			"giftValue": room.GiftValue, "heat": roomHeat(room), "supportRank": rank,
		}
		danmakulogic.GetHub(svcCtx).BroadcastEvent(roomID, "live_gift", payload)
		response.Ok(c, payload)
	}
}

func loadRecentSuperChats(svcCtx *svc.ServiceContext, room model.LiveRoom, limit int) ([]gin.H, error) {
	query := svcCtx.DB.Preload("User").Where("room_id = ? AND message <> ''", room.ID)
	if room.StartedAt != nil {
		query = query.Where("created_at >= ?", *room.StartedAt)
	}
	var gifts []model.LiveGift
	if err := query.Order("created_at DESC").Limit(limit).Find(&gifts).Error; err != nil {
		return nil, err
	}
	result := make([]gin.H, 0, len(gifts))
	for i := len(gifts) - 1; i >= 0; i-- {
		item := gifts[i]
		unitValue := int64(0)
		if item.Count > 0 {
			unitValue = item.Value / int64(item.Count)
		}
		result = append(result, gin.H{
			"id":    item.ID,
			"user":  &model.UserInfo{ID: item.User.ID, Username: item.User.Username, Nickname: item.User.Nickname, Avatar: item.User.Avatar, Role: item.User.Role},
			"gift":  giftDefinition{Key: item.GiftKey, Name: item.Name, Value: unitValue},
			"count": item.Count, "value": item.Value, "message": item.Message,
			"displaySeconds": item.DisplaySeconds, "createdAt": item.CreatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func superChatDisplaySeconds(value int64) int {
	switch {
	case value >= 1000:
		return 120
	case value >= 500:
		return 90
	case value >= 200:
		return 60
	case value >= 50:
		return 30
	default:
		return 15
	}
}

func loadSupportRank(svcCtx *svc.ServiceContext, roomID uint, limit int) ([]supportRankItem, error) {
	type row struct {
		UserID    uint
		Value     int64
		GiftCount int64
	}
	var rows []row
	err := svcCtx.DB.Model(&model.LiveGift{}).
		Select("user_id, SUM(value) AS value, SUM(count) AS gift_count").
		Where("room_id = ?", roomID).Group("user_id").Order("value DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]supportRankItem, 0, len(rows))
	for _, item := range rows {
		var user model.User
		if err := svcCtx.DB.First(&user, item.UserID).Error; err != nil {
			continue
		}
		result = append(result, supportRankItem{
			UserID: item.UserID, Value: item.Value, GiftCount: item.GiftCount,
			User: &model.UserInfo{ID: user.ID, Username: user.Username, Nickname: user.Nickname, Avatar: user.Avatar, Role: user.Role},
		})
	}
	return result, nil
}

func loadLiveRoom(c *gin.Context, svcCtx *svc.ServiceContext) (model.LiveRoom, bool) {
	roomID, err := parseRoomID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "无效的直播间 ID")
		return model.LiveRoom{}, false
	}
	var room model.LiveRoom
	if err := svcCtx.DB.Where("id = ? AND status = ?", roomID, "live").First(&room).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "直播间不存在或已结束")
		return model.LiveRoom{}, false
	}
	return room, true
}

func parseRoomID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid room id")
	}
	return uint(id), nil
}

func roomHeat(room model.LiveRoom) int64 {
	return room.ViewerCount*10 + room.LikeCount*2 + room.GiftValue
}
