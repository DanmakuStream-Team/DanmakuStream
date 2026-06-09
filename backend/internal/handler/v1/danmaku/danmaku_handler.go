package danmaku

import (
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type sendDanmakuReq struct {
	VideoID  uint   `json:"videoId" binding:"required"`
	Content  string `json:"content" binding:"required,max=500"`
	Time     int    `json:"time"`
	Color    string `json:"color"`
	FontSize string `json:"fontSize"`
	Type     string `json:"type"`
}

type advancedDanmakuUploadResult struct {
	List  []danmakuItem `json:"list"`
	Count int           `json:"count"`
}

type danmakuItem struct {
	ID       uint   `json:"id"`
	VideoID  uint   `json:"videoId"`
	UserID   uint   `json:"userId"`
	Content  string `json:"content"`
	Time     int    `json:"time"`
	Color    string `json:"color"`
	FontSize string `json:"fontSize"`
	Type     string `json:"type"`
}

type adminDanmakuListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	VideoID  uint   `form:"videoId"`
	Scene    string `form:"scene"`
	Keyword  string `form:"keyword"`
	Blocked  *bool  `form:"blocked"`
}

type adminDanmakuItem struct {
	ID        uint   `json:"id"`
	VideoID   uint   `json:"videoId"`
	Scene     string `json:"scene"`
	UserID    uint   `json:"userId"`
	Content   string `json:"content"`
	Time      int    `json:"time"`
	Color     string `json:"color"`
	FontSize  string `json:"fontSize"`
	Type      string `json:"type"`
	Blocked   bool   `json:"blocked"`
	CreatedAt string `json:"createdAt"`
}

type pageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

func ListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.ParseUint(c.Param("videoId"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var danmakus []model.Danmaku
		if err := svcCtx.DB.
			Where("video_id = ? AND scene = ? AND blocked = ?", videoID, "video", false).
			Order("time ASC").
			Find(&danmakus).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "弹幕加载失败")
			return
		}

		list := make([]danmakuItem, 0, len(danmakus))
		for _, d := range danmakus {
			list = append(list, danmakuItem{
				ID:       d.ID,
				VideoID:  d.VideoID,
				UserID:   d.UserID,
				Content:  d.Content,
				Time:     d.Time,
				Color:    d.Color,
				FontSize: d.FontSize,
				Type:     d.Type,
			})
		}
		response.Ok(c, list)
	}
}

func SendHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req sendDanmakuReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		if userID == 0 {
			response.Fail(c, http.StatusUnauthorized, "未授权")
			return
		}

		var video model.Video
		if err := svcCtx.DB.Select("id", "status").
			Where("id = ?", req.VideoID).
			First(&video).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}
		if video.Status != "approved" {
			response.Fail(c, http.StatusForbidden, "视频未通过审核")
			return
		}

		danmaku := model.Danmaku{
			VideoID:  req.VideoID,
			Scene:    "video",
			UserID:   userID,
			Content:  req.Content,
			Time:     req.Time,
			Color:    defaultStr(req.Color, "#FFFFFF"),
			FontSize: defaultStr(req.FontSize, "medium"),
			Type:     defaultStr(req.Type, "scroll"),
		}

		if err := svcCtx.DB.Create(&danmaku).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "弹幕发送失败")
			return
		}

		svcCtx.DB.Model(&model.Video{}).
			Where("id = ?", req.VideoID).
			UpdateColumn("danmaku_count", gorm.Expr("danmaku_count + ?", 1))

		response.Ok(c, danmakuItem{
			ID:       danmaku.ID,
			VideoID:  danmaku.VideoID,
			UserID:   danmaku.UserID,
			Content:  danmaku.Content,
			Time:     danmaku.Time,
			Color:    danmaku.Color,
			FontSize: danmaku.FontSize,
			Type:     danmaku.Type,
		})
	}
}

func UploadAdvancedHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		if userID == 0 {
			response.Fail(c, http.StatusUnauthorized, "未授权")
			return
		}

		videoID64, err := strconv.ParseUint(c.PostForm("videoId"), 10, 64)
		if err != nil || videoID64 == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}
		videoID := uint(videoID64)

		file, err := c.FormFile("file")
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "请选择 .danmaku 文件")
			return
		}
		if !strings.HasSuffix(strings.ToLower(file.Filename), ".danmaku") {
			response.Fail(c, http.StatusBadRequest, "文件后缀必须是 .danmaku")
			return
		}
		if file.Size > 512*1024 {
			response.Fail(c, http.StatusBadRequest, "弹幕文件不能超过 512KB")
			return
		}

		var video model.Video
		if err := svcCtx.DB.Select("id", "status").Where("id = ?", videoID).First(&video).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}
		if video.Status != "approved" {
			response.Fail(c, http.StatusForbidden, "视频未通过审核")
			return
		}

		src, err := file.Open()
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "文件读取失败")
			return
		}
		defer src.Close()

		buf, err := io.ReadAll(src)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "文件读取失败")
			return
		}

		danmakus, err := parseAdvancedDanmakuFile(string(buf), videoID, userID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		if len(danmakus) == 0 {
			response.Fail(c, http.StatusBadRequest, "弹幕文件没有有效内容")
			return
		}
		if len(danmakus) > 500 {
			response.Fail(c, http.StatusBadRequest, "单次最多上传 500 条高级弹幕")
			return
		}

		if err := svcCtx.DB.Create(&danmakus).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "高级弹幕创建失败")
			return
		}

		svcCtx.DB.Model(&model.Video{}).
			Where("id = ?", videoID).
			UpdateColumn("danmaku_count", gorm.Expr("danmaku_count + ?", len(danmakus)))

		list := make([]danmakuItem, 0, len(danmakus))
		for _, d := range danmakus {
			list = append(list, toDanmakuItem(d))
		}
		response.Ok(c, advancedDanmakuUploadResult{List: list, Count: len(list)})
	}
}

func AdminListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req adminDanmakuListReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}
		if req.PageSize > 100 {
			req.PageSize = 100
		}

		db := svcCtx.DB.Model(&model.Danmaku{})

		if req.VideoID > 0 {
			db = db.Where("video_id = ?", req.VideoID)
		}

		if req.Scene != "" {
			if req.Scene != "video" && req.Scene != "live" {
				response.Fail(c, http.StatusBadRequest, "无效的弹幕场景")
				return
			}
			db = db.Where("scene = ?", req.Scene)
		}

		if req.Keyword != "" {
			db = db.Where("content LIKE ?", "%"+req.Keyword+"%")
		}

		if req.Blocked != nil {
			db = db.Where("blocked = ?", *req.Blocked)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "弹幕列表加载失败")
			return
		}

		var danmakus []model.Danmaku
		if err := db.
			Order("created_at DESC").
			Offset((req.Page - 1) * req.PageSize).
			Limit(req.PageSize).
			Find(&danmakus).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "弹幕列表加载失败")
			return
		}

		list := make([]adminDanmakuItem, 0, len(danmakus))
		for _, d := range danmakus {
			list = append(list, adminDanmakuItem{
				ID:        d.ID,
				VideoID:   d.VideoID,
				Scene:     d.Scene,
				UserID:    d.UserID,
				Content:   d.Content,
				Time:      d.Time,
				Color:     d.Color,
				FontSize:  d.FontSize,
				Type:      d.Type,
				Blocked:   d.Blocked,
				CreatedAt: d.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		response.Ok(c, pageResult[adminDanmakuItem]{
			List:     list,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		})
	}
}

func BlockHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的弹幕 ID")
			return
		}

		result := svcCtx.DB.Model(&model.Danmaku{}).
			Where("id = ?", id).
			Update("blocked", true)
		if result.Error != nil {
			response.Fail(c, http.StatusInternalServerError, "屏蔽失败")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusNotFound, "弹幕不存在")
			return
		}
		response.Ok(c, nil)
	}
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func toDanmakuItem(d model.Danmaku) danmakuItem {
	return danmakuItem{
		ID:       d.ID,
		VideoID:  d.VideoID,
		UserID:   d.UserID,
		Content:  d.Content,
		Time:     d.Time,
		Color:    d.Color,
		FontSize: d.FontSize,
		Type:     d.Type,
	}
}

var advancedBlockRe = regexp.MustCompile(`\{([^{}]+)\}`)

func parseAdvancedDanmakuFile(content string, videoID, userID uint) ([]model.Danmaku, error) {
	blocks := advancedBlockRe.FindAllStringSubmatch(content, -1)
	danmakus := make([]model.Danmaku, 0, len(blocks))

	for _, block := range blocks {
		fields := parseAdvancedFields(block[1])
		text := strings.TrimSpace(fields["text"])
		if text == "" {
			text = "●"
		}

		timeValue, err := parseIntField(fields, "time", 0)
		if err != nil || timeValue < 0 {
			return nil, errAdvancedField("time")
		}

		x, err := parsePercentField(fields, "x", 50)
		if err != nil {
			return nil, errAdvancedField("x")
		}
		y, err := parsePercentField(fields, "y", 50)
		if err != nil {
			return nil, errAdvancedField("y")
		}
		tx, err := parsePercentField(fields, "tx", x)
		if err != nil {
			return nil, errAdvancedField("tx")
		}
		ty, err := parsePercentField(fields, "ty", y)
		if err != nil {
			return nil, errAdvancedField("ty")
		}
		dur, err := parseFloatField(fields, "dur", 4)
		if err != nil || dur <= 0 || dur > 30 {
			return nil, errAdvancedField("dur")
		}
		size, err := parseIntField(fields, "size", 24)
		if err != nil || size < 8 || size > 96 {
			return nil, errAdvancedField("size")
		}
		color := strings.TrimSpace(fields["color"])
		if color == "" {
			color = "#FFFFFF"
		}

		canonical := "@" + "adv " +
			"x=" + strconv.Itoa(x) + " " +
			"y=" + strconv.Itoa(y) + " " +
			"tx=" + strconv.Itoa(tx) + " " +
			"ty=" + strconv.Itoa(ty) + " " +
			"dur=" + trimFloat(dur) + " " +
			"size=" + strconv.Itoa(size) + " " +
			"color=" + color + " | " + text

		danmakus = append(danmakus, model.Danmaku{
			VideoID:  videoID,
			Scene:    "video",
			UserID:   userID,
			Content:  canonical,
			Time:     timeValue,
			Color:    color,
			FontSize: "medium",
			Type:     "advanced",
		})
	}

	return danmakus, nil
}

func parseAdvancedFields(block string) map[string]string {
	fields := make(map[string]string)
	for _, part := range strings.Split(block, ",") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" {
			fields[key] = value
		}
	}
	return fields
}

func parseIntField(fields map[string]string, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(fields[key])
	if raw == "" {
		return fallback, nil
	}
	return strconv.Atoi(raw)
}

func parseFloatField(fields map[string]string, key string, fallback float64) (float64, error) {
	raw := strings.TrimSpace(fields[key])
	if raw == "" {
		return fallback, nil
	}
	return strconv.ParseFloat(raw, 64)
}

func parsePercentField(fields map[string]string, key string, fallback int) (int, error) {
	value, err := parseIntField(fields, key, fallback)
	if err != nil {
		return 0, err
	}
	if value < 0 || value > 100 {
		return 0, errAdvancedField(key)
	}
	return value, nil
}

func trimFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(value, 'f', 2, 64), "0"), ".")
}

func errAdvancedField(name string) error {
	return &advancedFieldError{name: name}
}

type advancedFieldError struct {
	name string
}

func (e *advancedFieldError) Error() string {
	return "高级弹幕字段错误: " + e.name
}
