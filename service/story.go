package service

import (
	"errors"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Story struct{}

// Create 发布故事
func (s *Story) Create(photoID, userID int64, content, mediaURL string) (*model.Story, error) {
	// 检查图片是否存在
	var photo model.Photo
	if err := model.DB.First(&photo, photoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	story := &model.Story{
		PhotoID:  photoID,
		UserID:   userID,
		Content:  content,
		MediaURL: mediaURL,
		Likes:    0,
	}

	if err := model.DB.Create(story).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	// 加载用户信息
	model.DB.Preload("User").First(story, story.ID)

	return story, nil
}

// ListByPhoto 获取图片下的故事列表
func (s *Story) ListByPhoto(photoID int64) (map[string]any, error) {
	var stories []model.Story

	if err := model.DB.Where("photo_id = ?", photoID).
		Preload("User").
		Order("created_at DESC").
		Find(&stories).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	type StoryItem struct {
		ID        int64     `json:"id"`
		UserName  string    `json:"user_name"`
		Content   string    `json:"content"`
		MediaURL  string    `json:"media_url"`
		Likes     int       `json:"likes"`
		CreatedAt time.Time `json:"created_at"`
	}

	items := make([]StoryItem, 0, len(stories))
	for _, st := range stories {
		userName := ""
		if st.User.ID != 0 {
			userName = st.User.Name
		}
		items = append(items, StoryItem{
			ID:        st.ID,
			UserName:  userName,
			Content:   st.Content,
			MediaURL:  st.MediaURL,
			Likes:     st.Likes,
			CreatedAt: st.CreatedAt,
		})
	}

	return map[string]any{
		"stories": items,
	}, nil
}
