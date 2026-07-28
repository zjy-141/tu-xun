package service

import (
	"fmt"

	"tu-xun/model"
)

// ContentBlockSvc 内容位业务逻辑
type ContentBlockSvc struct{}

// GetByKey 获取内容位（默认值：content=""、version=0、updated_at=null）
func (s *ContentBlockSvc) GetByKey(key string) (*ContentBlock, error) {
	var cb model.ContentBlock
	err := model.DB.Where("`key` = ?", key).First(&cb).Error
	if err != nil {
		// 未编辑过的内容位返回默认值
		result := &ContentBlock{
			Key:       key,
			Content:   "",
			Version:   0,
			UpdatedAt: nil,
		}
		return result, nil
	}

	return &ContentBlock{
		Key:       cb.Key,
		Content:   cb.Content,
		Version:   cb.Version,
		UpdatedAt: &cb.UpdatedAt,
	}, nil
}

// AdminUpdate 管理端更新内容位（version 自增）
func (s *ContentBlockSvc) AdminUpdate(key string, req UpdateContentRequest) error {
	// 弹窗关联通知校验
	if key == "popup" && req.RelatedID > 0 {
		var count int64
		if err := model.DB.Model(&model.Announcement{}).Where("id = ?", req.RelatedID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("关联通知不存在")
		}
	}
	// 非 popup 不允许传 related_id
	if key != "popup" && req.RelatedID > 0 {
		return fmt.Errorf("该内容位不支持关联")
	}

	var cb model.ContentBlock
	err := model.DB.Where("`key` = ?", key).First(&cb).Error
	if err != nil {
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
		return model.DB.Create(&cb).Error
	}

	// 更新存在的内容位
	updates := map[string]interface{}{
		"content":  req.Content,
		"version":  cb.Version + 1,
	}
	if key == "popup" {
		updates["related_id"] = req.RelatedID
		if req.RelatedID > 0 {
			updates["related_type"] = "announcement"
		} else {
			updates["related_type"] = ""
		}
	}
	return model.DB.Model(&cb).Updates(updates).Error
}
