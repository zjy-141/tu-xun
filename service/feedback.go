package service

import (
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type FeedbackSvc struct {
}

// Create 发送反馈
func (f *FeedbackSvc) Create(info FeedbackCreateParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	feedback := &model.Feedback{
		UserID:  info.UserID,
		Title:   info.Title,
		Content: info.Content,
		Type:    info.Type,
		Phone:   info.Phone,
		Status:  0, // 0待处理
	}

	if err := tx.Create(feedback).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 如果有上传图片，保存图片并创建关联记录
	if info.ImageFile != nil {
		imageURL, _, err := saveUploadedFile(info.ImageFile, "photos")
		if err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}

		media := &model.FeedbackMedia{
			FeedbackID: uint(feedback.ID),
			URL:        imageURL,
			MediaType:  1, // 1图片
			Sort:       0,
		}
		if err := tx.Create(media).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = ResponseIS{
		ID:     feedback.ID,
		Status: "pending",
	}
	return resp, nil
}

// List 获取当前用户的反馈列表
func (f *FeedbackSvc) List(info FeedbackListParams) (resp FeedbackForms, err error) {
	var feedbacks []model.Feedback
	var total int64

	query := model.DB.Model(&model.Feedback{}).Where("user_id = ?", info.UserID)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at DESC").
		Scopes(model.Paginate(info.PagerForm)).
		Find(&feedbacks).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.Feedbacks = make([]FeedbackForm, 0, len(feedbacks))
	for _, fb := range feedbacks {
		resp.Feedbacks = append(resp.Feedbacks, FeedbackForm{
			ID:        fb.ID,
			Title:     fb.Title,
			Type:      fb.Type,
			Status:    fb.Status,
			CreatedAt: fb.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}

// Detail 获取反馈详情
func (f *FeedbackSvc) Detail(info FeedbackGetByIDParams) (resp FeedbackDetail, err error) {
	var feedback model.Feedback
	if err := model.DB.Preload("Medias", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).First(&feedback, info.FeedbackID).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	medias := make([]FeedbackMediaForm, 0, len(feedback.Medias))
	for _, m := range feedback.Medias {
		medias = append(medias, FeedbackMediaForm{
			ID:        m.ID,
			URL:       m.URL,
			MediaType: m.MediaType,
		})
	}

	resp = FeedbackDetail{
		ID:        feedback.ID,
		UserID:    feedback.UserID,
		Title:     feedback.Title,
		Content:   feedback.Content,
		Type:      feedback.Type,
		Phone:     feedback.Phone,
		Status:    feedback.Status,
		Medias:    medias,
		CreatedAt: feedback.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	return resp, nil
}
