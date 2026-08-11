package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"tu-xun/model"
	"tu-xun/pkg/htmlutil"
	"tu-xun/pkg/urlutil"

	"gorm.io/gorm"
)

const announcementContentPreviewLen = 50

// mediaPtr 当 originURL 为空时返回 nil，否则返回 *Media 指针
func mediaPtr(originURL string, width, height int) *Media {
	if originURL == "" {
		return nil
	}
	return &Media{OriginURL: originURL, Width: width, Height: height}
}

// AnnouncementSvc 通知/公告业务逻辑
type AnnouncementSvc struct{}

// generateContentPreview 生成正文摘要：剥离 HTML 标签（块级间补空格）→ 去 [image] 占位 → 合并空白 → 按 Unicode 码点截前 N 字
func generateContentPreview(content string, maxLen int) string {
	// 先剥离 HTML 标签，块级标签之间补空格防段落粘连
	s := htmlutil.StripHTML(content)
	s = strings.ReplaceAll(s, "[image]", " ")
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen])
}

// List 客户端通知列表（含未读数，登录用户）
func (s *AnnouncementSvc) List(userID int64, params AnnouncementListParams) (AnnouncementPage, error) {
	var total int64
	var list []AnnouncementListItem
	var unreadCount int64

	// 通知基础查询（按创建时间倒序，软删除过滤）
	q := model.DB.Model(&model.Announcement{}).Order("id DESC")

	// keyword 模糊匹配剥离标签后的正文文本
	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		q = q.Where("content_text LIKE ?", kw)
	}

	// 总数
	if err := q.Count(&total).Error; err != nil {
		return AnnouncementPage{}, err
	}

	// 分页查询
	rows, err := q.Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Rows()
	if err != nil {
		return AnnouncementPage{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var a model.Announcement
		if err := model.DB.ScanRows(rows, &a); err != nil {
			return AnnouncementPage{}, err
		}
		// 查当前用户是否已读
		var readCount int64
		model.DB.Model(&model.AnnouncementRead{}).
			Where("announcement_id = ? AND user_id = ?", a.ID, userID).
			Count(&readCount)
		list = append(list, AnnouncementListItem{
			ID:             a.ID,
			Title:          a.Title,
			ContentPreview: generateContentPreview(a.Content, announcementContentPreviewLen),
			IsRead:         readCount > 0,
			CreatedAt:      &a.CreatedAt,
		})
	}

	// 当前用户总未读数
	var totalAll, readAll int64
	model.DB.Model(&model.Announcement{}).Count(&totalAll)
	model.DB.Model(&model.AnnouncementRead{}).Where("user_id = ?", userID).Count(&readAll)
	unreadCount = totalAll - readAll
	if unreadCount < 0 {
		unreadCount = 0
	}

	return AnnouncementPage{
		Total:       total,
		List:        list,
		UnreadCount: unreadCount,
	}, nil
}

// GetByID 客户端通知详情（读取成功即标记已读）
func (s *AnnouncementSvc) GetByID(userID int64, id int64) (*AnnouncementDetail, error) {
	var a model.Announcement
	if err := model.DB.First(&a, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// 检查是否已读
	var readCount int64
	model.DB.Model(&model.AnnouncementRead{}).
		Where("announcement_id = ? AND user_id = ?", id, userID).
		Count(&readCount)
	isRead := readCount > 0

	// 写入已读记录（幂等）
	if !isRead {
		now := time.Now()
		model.DB.Create(&model.AnnouncementRead{
			AnnouncementID: id,
			UserID:         userID,
			ReadAt:         now,
		})
	}

	return &AnnouncementDetail{
		ID:          a.ID,
		Title:       a.Title,
		Content:     a.Content,
		Image:       mediaPtr(urlutil.FullURL(a.ImageURL), a.ImageWidth, a.ImageHeight),
		RelatedType: a.RelatedType,
		RelatedID:   a.RelatedID,
		IsRead:      isRead,
		CreatedAt:   &a.CreatedAt,
	}, nil
}

// AdminList 管理端通知列表（含已读人数，不按用户判断 is_read）
func (s *AnnouncementSvc) AdminList(params AdminAnnouncementListParams) (AdminAnnouncementPage, error) {
	var total int64
	var list []AdminAnnouncementListItem

	q := model.DB.Model(&model.Announcement{}).Order("id DESC")

	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		q = q.Where("id LIKE ? OR title LIKE ? OR content_text LIKE ?", kw, kw, kw)
	}

	if err := q.Count(&total).Error; err != nil {
		return AdminAnnouncementPage{}, err
	}

	rows, err := q.Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Rows()
	if err != nil {
		return AdminAnnouncementPage{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var a model.Announcement
		if err := model.DB.ScanRows(rows, &a); err != nil {
			return AdminAnnouncementPage{}, err
		}
		// 已读人数
		var readCount int64
		model.DB.Model(&model.AnnouncementRead{}).
			Where("announcement_id = ?", a.ID).
			Count(&readCount)
		list = append(list, AdminAnnouncementListItem{
			ID:             a.ID,
			Title:          a.Title,
			ContentPreview: generateContentPreview(a.Content, announcementContentPreviewLen),
			ReadCount:      readCount,
			CreatedAt:      &a.CreatedAt,
		})
	}

	return AdminAnnouncementPage{Total: total, List: list}, nil
}

// AdminGetByID 管理端通知详情（不标记已读）
func (s *AnnouncementSvc) AdminGetByID(id int64) (*AdminAnnouncementDetail, error) {
	var a model.Announcement
	if err := model.DB.First(&a, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	var readCount int64
	model.DB.Model(&model.AnnouncementRead{}).
		Where("announcement_id = ?", id).
		Count(&readCount)

	return &AdminAnnouncementDetail{
		ID:          a.ID,
		Title:       a.Title,
		Content:     a.Content,
		Image:       mediaPtr(urlutil.FullURL(a.ImageURL), a.ImageWidth, a.ImageHeight),
		RelatedType: a.RelatedType,
		RelatedID:   a.RelatedID,
		ReadCount:   readCount,
		CreatedAt:   &a.CreatedAt,
	}, nil
}

// Create 管理员发布通知
func (s *AnnouncementSvc) Create(params CreateAnnouncementRequest) (ResponseIS, error) {
	// 验证关联活动存在
	if params.RelatedType == "activity" && params.RelatedID > 0 {
		var count int64
		if err := model.DB.Model(&model.Activity{}).Where("id = ?", params.RelatedID).Count(&count).Error; err != nil {
			return ResponseIS{}, err
		}
		if count == 0 {
			return ResponseIS{}, fmt.Errorf("关联活动不存在")
		}
	}

	// HTML 白名单过滤 + 字数校验
	content := htmlutil.SanitizeHTML(params.Content)
	if err := htmlutil.ValidateRichText(content); err != nil {
		return ResponseIS{}, err
	}

	a := model.Announcement{
		SenderID:    1, // 系统/管理员
		Title:       params.Title,
		Content:     content,
		ContentText: htmlutil.PlainTextForSearch(content),
		RelatedType: params.RelatedType,
		RelatedID:   params.RelatedID,
	}

	// 上传配图
	if params.ImageFile != nil {
		uploadResult, err := saveUploadedFile(params.ImageFile, "announcements/images/", false)
		if err != nil {
			return ResponseIS{}, err
		}
		a.ImageURL = uploadResult.ImageURL
		a.ImageWidth = uploadResult.ImageWidth
		a.ImageHeight = uploadResult.ImageHeight
	}

	if err := model.DB.Create(&a).Error; err != nil {
		return ResponseIS{}, err
	}

	return ResponseIS{ID: a.ID, Status: "ok"}, nil
}

// Update 管理员更新通知
func (s *AnnouncementSvc) Update(id int64, params UpdateAnnouncementRequest) (ResponseIS, error) {
	var a model.Announcement
	if err := model.DB.First(&a, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ResponseIS{}, fmt.Errorf("通知不存在")
		}
		return ResponseIS{}, err
	}

	updates := map[string]interface{}{}

	if params.Title != "" {
		updates["title"] = params.Title
	}
	if params.Content != "" {
		content := htmlutil.SanitizeHTML(params.Content)
		if err := htmlutil.ValidateRichText(content); err != nil {
			return ResponseIS{}, err
		}
		updates["content"] = content
		updates["content_text"] = htmlutil.PlainTextForSearch(content)
	}

	// 处理 remove_image / remove_relation 与新值冲突
	if params.RemoveImage {
		if params.ImageFile != nil {
			return ResponseIS{}, fmt.Errorf("remove_image 与 image_file 不可同时传")
		}
		updates["image_url"] = ""
	} else if params.ImageFile != nil {
		uploadResult, err := saveUploadedFile(params.ImageFile, "announcements/images/", false)
		if err != nil {
			return ResponseIS{}, err
		}
		updates["image_url"] = uploadResult.ImageURL
		updates["image_width"] = uploadResult.ImageWidth
		updates["image_height"] = uploadResult.ImageHeight
	}

	if params.RemoveRelation {
		if params.RelatedType != "" || params.RelatedID > 0 {
			return ResponseIS{}, fmt.Errorf("remove_relation 与 related_type/related_id 不可同时传")
		}
		updates["related_type"] = ""
		updates["related_id"] = 0
	} else if params.RelatedType != "" {
		// 验证关联
		if params.RelatedType == "activity" && params.RelatedID > 0 {
			var count int64
			if err := model.DB.Model(&model.Activity{}).Where("id = ?", params.RelatedID).Count(&count).Error; err != nil {
				return ResponseIS{}, err
			}
			if count == 0 {
				return ResponseIS{}, fmt.Errorf("关联活动不存在")
			}
		}
		updates["related_type"] = params.RelatedType
		updates["related_id"] = params.RelatedID
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&a).Updates(updates).Error; err != nil {
			return ResponseIS{}, err
		}
	}

	return ResponseIS{ID: a.ID, Status: "ok"}, nil
}

// Delete 管理员删除通知
func (s *AnnouncementSvc) Delete(id int64) error {
	var a model.Announcement
	if err := model.DB.First(&a, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("通知不存在")
		}
		return err
	}
	return model.DB.Delete(&a).Error
}
