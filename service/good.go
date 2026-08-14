package service

import (
	"tu-xun/common"
	"tu-xun/model"
	"tu-xun/pkg/urlutil"
)

type GoodSvc struct{}

// List 获取上架奖品列表（仅 in_store）
func (s *GoodSvc) List(params GoodListParams) (GoodItemPage, error) {
	var goods []model.Good
	var total int64

	query := model.DB.Model(&model.Good{}).Where("status = ?", "in_store")

	if params.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+params.Keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return GoodItemPage{}, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("id DESC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&goods).Error; err != nil {
		return GoodItemPage{}, common.ErrNew(err, common.SysErr)
	}

	items := make([]GoodItem, 0, len(goods))
	for _, g := range goods {
		items = append(items, GoodItem{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Image: Media{
				OriginURL:   urlutil.FullURL(g.ImageURL),
				ThumbURL:    urlutil.FullURL(g.ThumbURL),
				Width:       g.ImageWidth,
				Height:      g.ImageHeight,
			},
			ScorePrice:  g.NeedScore,
			Stock:       g.Stock,
			Status:      g.Status,
			CreatedAt:   &g.CreatedAt,
		})
	}

	return GoodItemPage{
		Total: total,
		List:  items,
	}, nil
}
