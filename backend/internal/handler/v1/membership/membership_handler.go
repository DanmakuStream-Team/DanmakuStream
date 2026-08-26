package membership

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type planRequest struct {
	PriceCents int64  `json:"priceCents"`
	Benefits   string `json:"benefits"`
	Enabled    bool   `json:"enabled"`
}

type createOrderRequest struct {
	CreatorID uint `json:"creatorId"`
	Months    int  `json:"months"`
}

type autoRenewRequest struct {
	Enabled bool `json:"enabled"`
}

type planInfo struct {
	CreatorID  uint   `json:"creatorId"`
	PriceCents int64  `json:"priceCents"`
	Benefits   string `json:"benefits"`
	Enabled    bool   `json:"enabled"`
}

type subscriptionInfo struct {
	ID            uint           `json:"id"`
	Creator       model.UserInfo `json:"creator"`
	PriceCents    int64          `json:"priceCents"`
	Status        string         `json:"status"`
	AutoRenew     bool           `json:"autoRenew"`
	StartedAt     string         `json:"startedAt"`
	ExpiresAt     string         `json:"expiresAt"`
	DaysRemaining int            `json:"daysRemaining"`
}

type orderInfo struct {
	OrderNo     string         `json:"orderNo"`
	Creator     model.UserInfo `json:"creator"`
	AmountCents int64          `json:"amountCents"`
	Months      int            `json:"months"`
	Status      string         `json:"status"`
	PaidAt      *string        `json:"paidAt"`
	CreatedAt   string         `json:"createdAt"`
}

func PlanHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		creatorID, ok := parseID(c, "id")
		if !ok {
			return
		}
		var user model.User
		if err := svcCtx.DB.Select("id", "role").First(&user, creatorID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "创作者不存在")
			return
		}
		var plan model.CreatorMembershipPlan
		err := svcCtx.DB.Where("creator_id = ?", creatorID).First(&plan).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Ok(c, planInfo{CreatorID: creatorID, PriceCents: 500, Enabled: false})
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "订阅套餐加载失败")
			return
		}
		response.Ok(c, toPlanInfo(plan))
	}
}

func MyPlanHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		creatorID := c.GetUint(middleware.CtxKeyUserID)
		var plan model.CreatorMembershipPlan
		err := svcCtx.DB.Where("creator_id = ?", creatorID).First(&plan).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Ok(c, planInfo{CreatorID: creatorID, PriceCents: 500, Benefits: "特别关注标识、优先接收创作者动态", Enabled: false})
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "订阅套餐加载失败")
			return
		}
		response.Ok(c, toPlanInfo(plan))
	}
}

func UpdateMyPlanHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		creatorID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)
		if role != "creator" && role != "admin" {
			response.Fail(c, http.StatusForbidden, "只有创作者可以设置付费订阅")
			return
		}
		var req planRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		if req.PriceCents < 100 || req.PriceCents > 100000 {
			response.Fail(c, http.StatusBadRequest, "月费应在 1 元到 1000 元之间")
			return
		}
		benefits := strings.TrimSpace(req.Benefits)
		if len([]rune(benefits)) > 200 {
			response.Fail(c, http.StatusBadRequest, "权益说明不能超过 200 个字符")
			return
		}
		var plan model.CreatorMembershipPlan
		err := svcCtx.DB.Where("creator_id = ?", creatorID).First(&plan).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			plan = model.CreatorMembershipPlan{CreatorID: creatorID, PriceCents: req.PriceCents, Benefits: benefits, Enabled: req.Enabled}
			err = svcCtx.DB.Create(&plan).Error
		} else if err == nil {
			err = svcCtx.DB.Model(&plan).Updates(map[string]interface{}{
				"price_cents": req.PriceCents,
				"benefits":    benefits,
				"enabled":     req.Enabled,
			}).Error
			plan.PriceCents = req.PriceCents
			plan.Benefits = benefits
			plan.Enabled = req.Enabled
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "订阅套餐保存失败")
			return
		}
		response.Ok(c, toPlanInfo(plan))
	}
}

func StatusHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		subscriberID := c.GetUint(middleware.CtxKeyUserID)
		creatorID, ok := parseID(c, "id")
		if !ok {
			return
		}
		if err := expireUserSubscription(svcCtx, subscriberID, creatorID); err != nil {
			response.Fail(c, http.StatusInternalServerError, "订阅状态更新失败")
			return
		}
		var subscription model.CreatorSubscription
		err := svcCtx.DB.Preload("Creator").Where("subscriber_id = ? AND creator_id = ?", subscriberID, creatorID).First(&subscription).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Ok(c, gin.H{"active": false, "subscription": nil})
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "订阅状态加载失败")
			return
		}
		response.Ok(c, gin.H{"active": subscription.Status == "active" && subscription.ExpiresAt.After(time.Now()), "subscription": toSubscriptionInfo(subscription)})
	}
}

func CreateOrderHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		subscriberID := c.GetUint(middleware.CtxKeyUserID)
		var req createOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.CreatorID == 0 {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		if subscriberID == req.CreatorID {
			response.Fail(c, http.StatusBadRequest, "不能订阅自己")
			return
		}
		if req.Months != 1 && req.Months != 3 && req.Months != 12 {
			response.Fail(c, http.StatusBadRequest, "订阅时长仅支持 1、3 或 12 个月")
			return
		}
		var plan model.CreatorMembershipPlan
		if err := svcCtx.DB.Where("creator_id = ? AND enabled = ?", req.CreatorID, true).First(&plan).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, "该创作者暂未开放付费订阅")
			return
		}
		if blocked, err := hasBlockRelation(svcCtx.DB, subscriberID, req.CreatorID); err != nil || blocked {
			response.Fail(c, http.StatusForbidden, "存在拉黑关系，无法订阅")
			return
		}
		order := model.SubscriptionOrder{
			OrderNo: newOrderNo(), SubscriberID: subscriberID, CreatorID: req.CreatorID,
			AmountCents: plan.PriceCents * int64(req.Months), Months: req.Months, Status: "pending",
		}
		if err := svcCtx.DB.Create(&order).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "订阅订单创建失败")
			return
		}
		if err := svcCtx.DB.Preload("Creator").First(&order, order.ID).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "订阅订单加载失败")
			return
		}
		response.Ok(c, toOrderInfo(order))
	}
}

func DemoPayHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		subscriberID := c.GetUint(middleware.CtxKeyUserID)
		orderNo := strings.TrimSpace(c.Param("orderNo"))
		var result model.CreatorSubscription
		err := svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var order model.SubscriptionOrder
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ? AND subscriber_id = ?", orderNo, subscriberID).First(&order).Error; err != nil {
				return err
			}
			if order.Status == "paid" {
				return tx.Preload("Creator").Where("subscriber_id = ? AND creator_id = ?", subscriberID, order.CreatorID).First(&result).Error
			}
			if order.Status != "pending" {
				return errors.New("order is not payable")
			}
			if blocked, err := hasBlockRelation(tx, subscriberID, order.CreatorID); err != nil || blocked {
				if err != nil {
					return err
				}
				return errors.New("blocked relationship")
			}

			now := time.Now()
			var subscription model.CreatorSubscription
			err := tx.Where("subscriber_id = ? AND creator_id = ?", subscriberID, order.CreatorID).First(&subscription).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				subscription = model.CreatorSubscription{SubscriberID: subscriberID, CreatorID: order.CreatorID, StartedAt: now}
			} else if err != nil {
				return err
			}
			base := now
			if subscription.Status == "active" && subscription.ExpiresAt.After(now) {
				base = subscription.ExpiresAt
			}
			subscription.PriceCents = order.AmountCents / int64(order.Months)
			subscription.Status = "active"
			subscription.ExpiresAt = base.AddDate(0, order.Months, 0)
			if subscription.ID == 0 {
				if err := tx.Create(&subscription).Error; err != nil {
					return err
				}
			} else if err := tx.Save(&subscription).Error; err != nil {
				return err
			}

			if err := ensureFollow(tx, subscriberID, order.CreatorID); err != nil {
				return err
			}
			if err := tx.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ?", subscriberID, order.CreatorID).Update("special", true).Error; err != nil {
				return err
			}
			paidAt := now
			if err := tx.Model(&order).Updates(map[string]interface{}{"status": "paid", "paid_at": &paidAt}).Error; err != nil {
				return err
			}
			actorID := subscriberID
			if err := tx.Create(&model.Notification{UserID: order.CreatorID, ActorID: &actorID, Type: "membership", Title: "收到新的付费特别关注", Content: fmt.Sprintf("订阅 %d 个月", order.Months), Link: "/creator"}).Error; err != nil {
				return err
			}
			return tx.Preload("Creator").First(&result, subscription.ID).Error
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "订阅订单不存在")
			return
		}
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "订单无法支付")
			return
		}
		response.Ok(c, toSubscriptionInfo(result))
	}
}

func MineHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		_ = expireAll(svcCtx)
		var subscriptions []model.CreatorSubscription
		if err := svcCtx.DB.Preload("Creator").Where("subscriber_id = ?", userID).Order("expires_at DESC").Find(&subscriptions).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "付费订阅加载失败")
			return
		}
		list := make([]subscriptionInfo, 0, len(subscriptions))
		for _, item := range subscriptions {
			list = append(list, toSubscriptionInfo(item))
		}
		response.Ok(c, gin.H{"list": list})
	}
}

func OrderListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		var orders []model.SubscriptionOrder
		if err := svcCtx.DB.Preload("Creator").Where("subscriber_id = ?", userID).Order("created_at DESC").Limit(100).Find(&orders).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "订阅订单加载失败")
			return
		}
		list := make([]orderInfo, 0, len(orders))
		for _, item := range orders {
			list = append(list, toOrderInfo(item))
		}
		response.Ok(c, gin.H{"list": list})
	}
}

func AutoRenewHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		creatorID, ok := parseID(c, "creatorId")
		if !ok {
			return
		}
		var req autoRenewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		if req.Enabled {
			response.Fail(c, http.StatusBadRequest, "演示支付暂不支持自动续费")
			return
		}
		result := svcCtx.DB.Model(&model.CreatorSubscription{}).
			Where("subscriber_id = ? AND creator_id = ? AND status = ?", userID, creatorID, "active").Update("auto_renew", req.Enabled)
		if result.Error != nil {
			response.Fail(c, http.StatusInternalServerError, "续费设置保存失败")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusNotFound, "有效订阅不存在")
			return
		}
		response.Ok(c, gin.H{"autoRenew": req.Enabled})
	}
}

func StartExpirationWorker(svcCtx *svc.ServiceContext) {
	_ = expireAll(svcCtx)
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			_ = expireAll(svcCtx)
		}
	}()
}

func expireAll(svcCtx *svc.ServiceContext) error {
	return svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var expired []model.CreatorSubscription
		if err := tx.Where("status = ? AND expires_at <= ?", "active", time.Now()).Find(&expired).Error; err != nil {
			return err
		}
		for _, item := range expired {
			if err := tx.Model(&item).Updates(map[string]interface{}{"status": "expired", "auto_renew": false}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ?", item.SubscriberID, item.CreatorID).Update("special", false).Error; err != nil {
				return err
			}
		}
		var specialFollows []model.Follow
		if err := tx.Where("special = ?", true).Find(&specialFollows).Error; err != nil {
			return err
		}
		for _, follow := range specialFollows {
			var count int64
			if err := tx.Model(&model.CreatorSubscription{}).Where("subscriber_id = ? AND creator_id = ? AND status = ? AND expires_at > ?", follow.FollowerID, follow.FolloweeID, "active", time.Now()).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				if err := tx.Model(&follow).Update("special", false).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func expireUserSubscription(svcCtx *svc.ServiceContext, subscriberID, creatorID uint) error {
	var item model.CreatorSubscription
	err := svcCtx.DB.Where("subscriber_id = ? AND creator_id = ?", subscriberID, creatorID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil || item.Status != "active" || item.ExpiresAt.After(time.Now()) {
		return err
	}
	return svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&item).Updates(map[string]interface{}{"status": "expired", "auto_renew": false}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ?", subscriberID, creatorID).Update("special", false).Error
	})
}

func ensureFollow(tx *gorm.DB, followerID, followeeID uint) error {
	var count int64
	if err := tx.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if err := tx.Create(&model.Follow{FollowerID: followerID, FolloweeID: followeeID}).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.User{}).Where("id = ?", followerID).UpdateColumn("follow_count", gorm.Expr("follow_count + 1")).Error; err != nil {
		return err
	}
	return tx.Model(&model.User{}).Where("id = ?", followeeID).UpdateColumn("fan_count", gorm.Expr("fan_count + 1")).Error
}

func hasBlockRelation(db *gorm.DB, firstID, secondID uint) (bool, error) {
	var count int64
	err := db.Model(&model.UserBlock{}).Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", firstID, secondID, secondID, firstID).Count(&count).Error
	return count > 0, err
}

func toPlanInfo(plan model.CreatorMembershipPlan) planInfo {
	return planInfo{CreatorID: plan.CreatorID, PriceCents: plan.PriceCents, Benefits: plan.Benefits, Enabled: plan.Enabled}
}

func toSubscriptionInfo(item model.CreatorSubscription) subscriptionInfo {
	days := int(math.Ceil(time.Until(item.ExpiresAt).Hours() / 24))
	if days < 0 {
		days = 0
	}
	return subscriptionInfo{ID: item.ID, Creator: userInfo(item.Creator), PriceCents: item.PriceCents, Status: item.Status, AutoRenew: item.AutoRenew, StartedAt: item.StartedAt.Format("2006-01-02 15:04:05"), ExpiresAt: item.ExpiresAt.Format("2006-01-02 15:04:05"), DaysRemaining: days}
}

func toOrderInfo(item model.SubscriptionOrder) orderInfo {
	var paidAt *string
	if item.PaidAt != nil {
		value := item.PaidAt.Format("2006-01-02 15:04:05")
		paidAt = &value
	}
	return orderInfo{OrderNo: item.OrderNo, Creator: userInfo(item.Creator), AmountCents: item.AmountCents, Months: item.Months, Status: item.Status, PaidAt: paidAt, CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05")}
}

func userInfo(user model.User) model.UserInfo {
	return model.UserInfo{ID: user.ID, Username: user.Username, Nickname: user.Nickname, Avatar: user.Avatar, Role: user.Role}
}

func newOrderNo() string {
	buffer := make([]byte, 5)
	_, _ = rand.Read(buffer)
	return fmt.Sprintf("DS%d%s", time.Now().UnixMilli(), strings.ToUpper(hex.EncodeToString(buffer)))
}

func parseID(c *gin.Context, key string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || value == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的 ID")
		return 0, false
	}
	return uint(value), true
}
