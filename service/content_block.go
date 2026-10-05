package service

import (
	"errors"

	"tu-xun/common"
	"tu-xun/model"
	"tu-xun/pkg/htmlutil"

	"gorm.io/gorm"
)

// ContentBlockSvc 内容位业务逻辑
type ContentBlockSvc struct{}

// GetByKey 获取内容位（默认值：content=""、version=0、updated_at=null）
func (s *ContentBlockSvc) GetByKey(key string) (resp *ContentBlock, err error) {
	var cb model.ContentBlock
	if err := model.DB.Where("`key` = ?", key).First(&cb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 未编辑过的内容位返回默认值
			return &ContentBlock{
				Key:       key,
				Content:   "",
				Version:   0,
				UpdatedAt: nil,
			}, nil
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	return &ContentBlock{
		Key:       cb.Key,
		Content:   cb.Content,
		RelatedID: cb.RelatedID,
		Version:   cb.Version,
		UpdatedAt: &cb.UpdatedAt,
	}, nil
}

// AdminUpdate 管理端更新内容位（version 自增）
func (s *ContentBlockSvc) AdminUpdate(key string, req UpdateContentRequest) (err error) {
	// HTML 白名单过滤 + 字数校验
	content := htmlutil.SanitizeHTML(req.Content)
	if err := htmlutil.ValidateRichText(content); err != nil {
		return common.ErrNew(err, common.ParamErr)
	}
	req.Content = content

	// 弹窗关联通知校验
	if key == "popup" && req.RelatedID > 0 {
		var count int64
		if err := model.DB.Model(&model.Announcement{}).Where("id = ?", req.RelatedID).Count(&count).Error; err != nil {
			return common.ErrNew(err, common.SysErr)
		}
		if count == 0 {
			return common.ErrNew(errors.New("关联通知不存在"), common.OpErr)
		}
	}
	// 非 popup 不允许传 related_id
	if key != "popup" && req.RelatedID > 0 {
		return common.ErrNew(errors.New("该内容位不支持关联"), common.ParamErr)
	}

	var cb model.ContentBlock
	if err := model.DB.Where("`key` = ?", key).First(&cb).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNew(err, common.SysErr)
		}
		// 首次创建
		cb = model.ContentBlock{
			Key:         key,
			Content:     req.Content,
			Version:     1,
			RelatedID:   req.RelatedID,
			RelatedType: "",
		}
		if key == "popup" && req.RelatedID > 0 {
			cb.RelatedType = "announcement"
		}
		if err := model.DB.Create(&cb).Error; err != nil {
			return common.ErrNew(err, common.SysErr)
		}
		return nil
	}

	// 更新存在的内容位
	updates := map[string]interface{}{
		"content": req.Content,
		"version": cb.Version + 1,
	}
	if key == "popup" {
		updates["related_id"] = req.RelatedID
		if req.RelatedID > 0 {
			updates["related_type"] = "announcement"
		} else {
			updates["related_type"] = ""
		}
	}
	if err := model.DB.Model(&cb).Updates(updates).Error; err != nil {
		return common.ErrNew(err, common.SysErr)
	}
	return nil
}
