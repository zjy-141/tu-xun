package service

import (
	"tu-xun/common"
	"tu-xun/model"
)

type Prize struct{}

// MyPrizes 获取我的奖品列表
func (info *Prize) MyPrizes(userID int64) (resp MyPrizesResponse, err error) {
	var prizes []model.Prize

	if err := model.DB.Where("user_id = ?", userID).
		Preload("Photo").
		Order("awarded_at DESC").
		Find(&prizes).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
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

	resp = MyPrizesResponse{
		Prizes: items,
	}
	return resp, nil
}
