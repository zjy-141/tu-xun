package service

import (
	"errors"
	"mime/multipart"

	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type FeedbackSvc struct{}

// Create 发送反馈
func (f *FeedbackSvc) Create(params FeedbackCreateParams) (ResponseIS, error) {
	tx := model.DB.Begin()
	var err error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	feedback := &model.Feedback{
		UserID:  params.UserID,
		Title:   params.Title,
		Content: params.Content,
		Type:    params.Type,
		Phone:   params.Phone,
		Status:  "pending",
	}

	if err = tx.Create(feedback).Error; err != nil {
		return ResponseIS{}, common.ErrNew(err, common.SysErr)
	}

	files := []*multipart.FileHeader{params.MediaFile1, params.MediaFile2, params.MediaFile3}
	for i, file := range files {
		if file == nil {
			continue
		}
		imageURL, _, err := saveUploadedFile(file, "feedbacks", false)
		if err != nil {
			return ResponseIS{}, common.ErrNew(err, common.SysErr)
		}
		media := &model.FeedbackMedia{
			FeedbackID: feedback.ID,
			URL:        imageURL,
			MediaType:  1,
			Sort:       i + 1,
		}
		if err = tx.Create(media).Error; err != nil {
			return ResponseIS{}, common.ErrNew(err, common.SysErr)
		}
	}

	if err = tx.Commit().Error; err != nil {
		return ResponseIS{}, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return ResponseIS{
		ID:     feedback.ID,
		Status: "pending",
	}, nil
}

// List 获取反馈列表
func (f *FeedbackSvc) List(params FeedbackListParams) (FeedbackPage, error) {
	var feedbacks []model.Feedback
	var total int64

	query := model.DB.Model(&model.Feedback{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Type > 0 {
		query = query.Where("type = ?", params.Type)
	}
	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", kw, kw)
	}

	if err := query.Count(&total).Error; err != nil {
		return FeedbackPage{}, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("User").
		Order("created_at DESC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&feedbacks).Error; err != nil {
		return FeedbackPage{}, common.ErrNew(err, common.SysErr)
	}

	resp := FeedbackPage{
		Total: total,
		List:  make([]FeedbackItem, 0, len(feedbacks)),
	}
	for _, fb := range feedbacks {
		resp.List = append(resp.List, FeedbackItem{
			ID: fb.ID,
			User: UserBrief{
				ID:        fb.User.ID,
				Nickname:  fb.User.Nickname,
				AvatarURL: fb.User.AvatarURL,
			},
			Title:     fb.Title,
			Type:      fb.Type,
			Status:    fb.Status,
			CreatedAt: &fb.CreatedAt,
		})
	}
	return resp, nil
}

// Detail 获取反馈详情
func (f *FeedbackSvc) Detail(feedbackID int64) (*FeedbackDetail, error) {
	var feedback model.Feedback
	if err := model.DB.Preload("Medias", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).Preload("User").First(&feedback, feedbackID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("反馈不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	medias := make([]FeedbackMediaItem, 0, len(feedback.Medias))
	for _, m := range feedback.Medias {
		medias = append(medias, FeedbackMediaItem{
			ID:        m.ID,
			URL:       m.URL,
			MediaType: m.MediaType,
		})
	}

	return &FeedbackDetail{
		ID: feedback.ID,
		User: UserBrief{
			ID:        feedback.User.ID,
			Nickname:  feedback.User.Nickname,
			AvatarURL: feedback.User.AvatarURL,
		},
		Title:     feedback.Title,
		Content:   feedback.Content,
		Type:      feedback.Type,
		Phone:     feedback.Phone,
		Status:    feedback.Status,
		Medias:    medias,
		CreatedAt: &feedback.CreatedAt,
	}, nil
}

// Review 回复反馈（更新状态）
func (f *FeedbackSvc) Review(params FeedbackReviewParams) error {
	tx := model.DB.Begin()
	var err error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	result := tx.Model(&model.Feedback{}).
		Where("id = ?", params.FeedbackID).
		Update("status", params.Status)

	if result.Error != nil {
		return common.ErrNew(result.Error, common.SysErr)
	}
	if result.RowsAffected == 0 {
		return common.ErrNew(errors.New("没有找到该反馈"), common.OpErr)
	}

	if err = tx.Commit().Error; err != nil {
		return common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return nil
}
