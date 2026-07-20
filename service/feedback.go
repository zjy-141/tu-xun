package service

import (
	"errors"
	"mime/multipart"
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
		Status:  "pending",
	}

	if err := tx.Create(feedback).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	files := []*multipart.FileHeader{info.ImageFile1, info.ImageFile2, info.ImageFile3}
	for i, file := range files {
		if file == nil {
			continue
		}
		imageURL, _, err := saveUploadedFile(file, "feedbacks", false)
		if err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
		media := &model.FeedbackMedia{
			FeedbackID: feedback.ID,
			URL:        imageURL,
			MediaType:  1,
			Sort:       i + 1, // 按照在切片中的位置赋值
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

	query := model.DB.Model(&model.Feedback{})

	if info.Status != "" {
		query = query.Where("status = ?", info.Status)
	}
	if info.Type > 0 {
		query = query.Where("type = ?", info.Type)
	}
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
			CreatedAt: &fb.CreatedAt,
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
		CreatedAt: &feedback.CreatedAt,
	}
	return resp, nil
}

// Review 回复反馈（更新状态）
func (f *FeedbackSvc) Review(info FeedbackReviewParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	result := tx.Model(&model.Feedback{}).
		Where("id = ?", info.FeedbackID).
		Update("status", info.Status)

	// 处理更新错误
	if result.Error != nil {
		tx.Rollback()
		return resp, common.ErrNew(result.Error, common.SysErr)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("没有找到该反馈"), common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = ResponseIS{
		ID:     info.FeedbackID,
		Status: info.Status,
	}
	return resp, nil
}
