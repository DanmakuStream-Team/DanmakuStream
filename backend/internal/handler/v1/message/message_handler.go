package message

import (
	"errors"
	"net/http"
	"sort"
	"strconv"

	"danmakustream/backend/internal/handler/response"
	chatlogic "danmakustream/backend/internal/logic/chat"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
)

type conversationInfo struct {
	User        model.UserInfo        `json:"user"`
	LastMessage chatlogic.MessageInfo `json:"lastMessage"`
	UnreadCount int64                 `json:"unreadCount"`
}

func ConversationListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		var messages []model.ChatMessage
		if err := svcCtx.DB.Preload("Sender").Preload("Receiver").Preload("SharedVideo.Author").
			Where("sender_id = ? OR receiver_id = ?", userID, userID).
			Order("created_at DESC").
			Limit(5000).
			Find(&messages).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "会话列表加载失败")
			return
		}

		otherIDs := make([]uint, 0)
		seenIDs := make(map[uint]struct{})
		for _, item := range messages {
			otherID := item.SenderID
			if otherID == userID {
				otherID = item.ReceiverID
			}
			if _, exists := seenIDs[otherID]; !exists {
				seenIDs[otherID] = struct{}{}
				otherIDs = append(otherIDs, otherID)
			}
		}

		var users []model.User
		if len(otherIDs) > 0 {
			if err := svcCtx.DB.Where("id IN ?", otherIDs).Find(&users).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "会话用户加载失败")
				return
			}
		}
		userMap := make(map[uint]model.User, len(users))
		for _, item := range users {
			userMap[item.ID] = item
		}

		conversationMap := make(map[uint]*conversationInfo)
		for _, item := range messages {
			otherID := item.SenderID
			if otherID == userID {
				otherID = item.ReceiverID
			}
			conversation := conversationMap[otherID]
			if conversation == nil {
				other := userMap[otherID]
				conversation = &conversationInfo{
					User:        model.UserInfo{ID: other.ID, Username: other.Username, Nickname: other.Nickname, Avatar: other.Avatar, Role: other.Role},
					LastMessage: chatlogic.ToMessageInfo(item),
				}
				conversationMap[otherID] = conversation
			}
			if item.ReceiverID == userID && !item.Read {
				conversation.UnreadCount++
			}
		}

		list := make([]conversationInfo, 0, len(conversationMap))
		for _, conversation := range conversationMap {
			list = append(list, *conversation)
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].LastMessage.ID > list[j].LastMessage.ID
		})
		response.Ok(c, gin.H{"list": list})
	}
}

func HistoryHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		otherID, ok := parseUserID(c)
		if !ok {
			return
		}
		page, pageSize := getPage(c)
		db := svcCtx.DB.Model(&model.ChatMessage{}).
			Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, otherID, otherID, userID)
		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "消息历史加载失败")
			return
		}
		var messages []model.ChatMessage
		if err := db.Preload("Sender").Preload("Receiver").Preload("SharedVideo.Author").Order("created_at DESC").
			Offset((page - 1) * pageSize).Limit(pageSize).Find(&messages).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "消息历史加载失败")
			return
		}
		if err := markConversationRead(svcCtx, userID, otherID); err != nil {
			response.Fail(c, http.StatusInternalServerError, "消息已读状态更新失败")
			return
		}
		list := make([]chatlogic.MessageInfo, 0, len(messages))
		for index := len(messages) - 1; index >= 0; index-- {
			if messages[index].ReceiverID == userID {
				messages[index].Read = true
			}
			list = append(list, chatlogic.ToMessageInfo(messages[index]))
		}
		response.Ok(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
	}
}

func SendHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		var req chatlogic.CreateMessageInput
		if err := c.ShouldBindJSON(&req); err != nil || req.ReceiverID == 0 {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		message, err := chatlogic.GetHub(svcCtx).CreateAndBroadcast(userID, req)
		if err != nil {
			writeSendError(c, err)
			return
		}
		response.Ok(c, message)
	}
}

func ReadHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		otherID, ok := parseUserID(c)
		if !ok {
			return
		}
		if err := markConversationRead(svcCtx, userID, otherID); err != nil {
			response.Fail(c, http.StatusInternalServerError, "消息已读状态更新失败")
			return
		}
		response.Ok(c, nil)
	}
}

func UnreadHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		var count int64
		if err := svcCtx.DB.Model(&model.ChatMessage{}).
			Where("receiver_id = ? AND `read` = ?", userID, false).
			Count(&count).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "未读私信数量加载失败")
			return
		}
		response.Ok(c, gin.H{"count": count})
	}
}

func markConversationRead(svcCtx *svc.ServiceContext, userID, otherID uint) error {
	return svcCtx.DB.Model(&model.ChatMessage{}).
		Where("sender_id = ? AND receiver_id = ? AND `read` = ?", otherID, userID, false).
		Update("read", true).Error
}

func parseUserID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil || value == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的用户 ID")
		return 0, false
	}
	return uint(value), true
}

func getPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	return page, pageSize
}

func writeSendError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, chatlogic.ErrBlocked) {
		status = http.StatusForbidden
	} else if errors.Is(err, chatlogic.ErrUserNotFound) {
		status = http.StatusNotFound
	}
	response.Fail(c, status, chatlogic.ErrorMessage(err))
}
