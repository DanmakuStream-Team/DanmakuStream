package analytics

import (
	"time"

	model "danmakustream/backend/internal/model/mysql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AddCreatorDailyStat(db *gorm.DB, creatorID uint, viewDelta, collectDelta, streamCount int64) error {
	if creatorID == 0 || (viewDelta == 0 && collectDelta == 0 && streamCount == 0) {
		return nil
	}

	stat := model.CreatorDailyStat{
		CreatorID:    creatorID,
		Date:         time.Now().Format("2006-01-02"),
		ViewDelta:    viewDelta,
		CollectDelta: collectDelta,
		StreamCount:  streamCount,
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "creator_id"}, {Name: "date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"view_delta":    gorm.Expr("view_delta + ?", viewDelta),
			"collect_delta": gorm.Expr("collect_delta + ?", collectDelta),
			"stream_count":  gorm.Expr("stream_count + ?", streamCount),
		}),
	}).Create(&stat).Error
}

func AddVideoDailyStat(db *gorm.DB, creatorID, videoID uint, viewDelta, collectDelta int64) error {
	if creatorID == 0 || videoID == 0 || (viewDelta == 0 && collectDelta == 0) {
		return nil
	}

	stat := model.VideoDailyStat{
		CreatorID:    creatorID,
		VideoID:      videoID,
		Date:         time.Now().Format("2006-01-02"),
		ViewDelta:    viewDelta,
		CollectDelta: collectDelta,
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "video_id"}, {Name: "date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"view_delta":    gorm.Expr("view_delta + ?", viewDelta),
			"collect_delta": gorm.Expr("collect_delta + ?", collectDelta),
		}),
	}).Create(&stat).Error
}
