package service

import (
	"time"
	"tu-xun/common"
	"tu-xun/model"
)

type Prize struct{}

// MyPrizes 获取我的奖品列表
func (info *Prize) MyPrizes(userID int64) (map[string]any, error) {
	var prizes []model.Prize

	if err := model.DB.Where("user_id = ?", userID).
		Preload("Photo").
		Order("awarded_at DESC").
		Find(&prizes).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	type PrizeItem struct {
		ID         int64      `json:"id"`
		PhotoID    int64      `json:"photo_id"`
		PhotoTitle string     `json:"photo_title"`
		Status     string     `json:"status"`
		PrizeType  string     `json:"prize_type"`
		AwardedAt  *time.Time `json:"awarded_at"`
	}

	items := make([]PrizeItem, 0, len(prizes))
	for _, pz := range prizes {
		photoTitle := ""
		if pz.Photo.ID != 0 {
			photoTitle = pz.Photo.Title
		}
		items = append(items, PrizeItem{
			ID:         pz.ID,
			PhotoID:    pz.PhotoID,
			PhotoTitle: photoTitle,
			Status:     pz.Status,
			PrizeType:  pz.PrizeType,
			AwardedAt:  pz.AwardedAt,
		})
	}

	return map[string]any{
		"prizes": items,
	}, nil
}
