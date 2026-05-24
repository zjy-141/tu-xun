package service

import (
	"tu-xun/common"
	"tu-xun/model"
)

type Prize struct{}

// MyPrizes 获取我的奖品列表
func (info *Prize) MyPrizes(params MyPrizesParams) (resp MyPrizesResponse, err error) {
	var prizes []model.Prize
	var total int64

	query := model.DB.Model(&model.Prize{}).
		Where("user_id = ?", params.UserID)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	if err := query.Scopes(model.Paginate(params.PagerForm)).
		Find(&prizes).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Prizes = make([]PrizeForm, 0, len(prizes))
	for _, pz := range prizes {
		resp.Prizes = append(resp.Prizes, PrizeForm{
			ID:         pz.ID,
			PhotoID:    pz.PhotoID,
			PhotoTitle: pz.Photo.Title,
			Status:     pz.Status,
			PrizeType:  pz.PrizeType,
			AwardedAt:  pz.AwardedAt,
		})
	}

	return resp, nil
}
