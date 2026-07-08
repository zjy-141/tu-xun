package model

import (
	"tu-xun/common"

	"gorm.io/gorm"
)

// Paginate 返回 GORM Scope 函数，按分页参数设置 Offset/Limit
func Paginate(pager common.PagerForm) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pager.Page <= 0 {
			pager.Page = 1
		}

		switch {
		case pager.PageSize > 20:
			pager.PageSize = 20
		case pager.PageSize <= 0:
			pager.PageSize = 10
		}

		offset := (pager.Page - 1) * pager.PageSize
		return db.Offset(offset).Limit(pager.PageSize)
	}
}
