package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"
	"tu-xun/pkg/urlutil"

	"gorm.io/gorm"
)

type FeedbackSvc struct{}

// Create 发送反馈（支持单个 media_file：图片≤20MB 或视频≤50MB）
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

	// 处理单个附件（图片或视频）
	if params.MediaFile != nil {
		originURL, thumbURL, width, height, mediaType, uploadErr := saveUploadedMedia(params.MediaFile, "feedbacks")
		if uploadErr != nil {
			tx.Rollback()
			return ResponseIS{}, common.ErrNew(uploadErr, common.SysErr)
		}
		media := &model.FeedbackMedia{
			FeedbackID: feedback.ID,
			URL:        originURL,
			ThumbURL:   thumbURL,
			Width:      width,
			Height:     height,
			MediaType:  mediaType,
			Sort:       1,
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
				ID:       fb.User.ID,
				Nickname: fb.User.Nickname,
				Avatar:   fb.User.AvatarURL,
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

	medias := make([]FeedbackMedia, 0, len(feedback.Medias))
	for _, m := range feedback.Medias {
		fm := FeedbackMedia{
			ID:        m.ID,
			OriginURL: urlutil.FullURL(m.URL),
			MediaType: m.MediaType,
		}
		if m.ThumbURL != "" {
			fm.ThumbURL = urlutil.FullURL(m.ThumbURL)
		}
		if m.Width > 0 {
			w := m.Width
			fm.Width = &w
		}
		if m.Height > 0 {
			h := m.Height
			fm.Height = &h
		}
		medias = append(medias, fm)
	}

	return &FeedbackDetail{
		ID: feedback.ID,
		User: UserBrief{
			ID:       feedback.User.ID,
			Nickname: feedback.User.Nickname,
			Avatar:   feedback.User.AvatarURL,
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
		tx.Rollback()
		return common.ErrNew(result.Error, common.SysErr)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return common.ErrNew(errors.New("没有找到该反馈"), common.OpErr)
	}

	if err = tx.Commit().Error; err != nil {
		return common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return nil
}
