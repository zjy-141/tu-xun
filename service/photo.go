package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Photo struct{}

// Create 上传图片投稿
func (p *Photo) Create(params CreatePhotoParams) (*model.Photo, error) {
	// 保存图片
	imageURL, thumbURL, err := saveUploadedFile(params.ImageFile, "photos")
	if err != nil {
		return nil, err
	}

	photo := &model.Photo{
		UserID:         params.UserID,
		Title:          params.Title,
		Description:    params.Description,
		ImageURL:       imageURL,
		ThumbURL:       thumbURL,
		LocationSecret: params.LocationSecret,
		Status:         "pending",
		Solved:         false,
		AttemptsCount:  0,
	}

	if err := model.DB.Create(photo).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	return photo, nil
}

// List 获取已审核通过的图片列表
func (p *Photo) List(params ListPhotoParams) (map[string]any, error) {
	var photos []model.Photo
	var total int64

	query := model.DB.Model(&model.Photo{}).Where("status = ?", "approved")

	if params.Solved != nil {
		query = query.Where("solved = ?", *params.Solved)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("Author").
		Order("created_at DESC").
		Scopes(model.Paginate(common.PagerForm{Page: params.Page, Limit: params.Limit})).
		Find(&photos).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	// 隐藏敏感字段
	type PhotoItem struct {
		ID            int64     `json:"id"`
		Title         string    `json:"title"`
		Description   string    `json:"description"`
		ImageURL      string    `json:"image_url"`
		Author        UserBrief `json:"author"`
		Solved        bool      `json:"solved"`
		AttemptsCount int       `json:"attempts_count"`
		CreatedAt     time.Time `json:"created_at"`
	}

	items := make([]PhotoItem, 0, len(photos))
	for _, ph := range photos {
		items = append(items, PhotoItem{
			ID:            ph.ID,
			Title:         ph.Title,
			Description:   ph.Description,
			ImageURL:      ph.ThumbURL,
			Author:        UserBrief{ID: ph.Author.ID, Name: ph.Author.Name},
			Solved:        ph.Solved,
			AttemptsCount: ph.AttemptsCount,
			CreatedAt:     ph.CreatedAt,
		})
	}

	return map[string]any{
		"total": total,
		"page":  params.Page,
		"limit": params.Limit,
		"items": items,
	}, nil
}

// GetByID 获取图片详情
func (p *Photo) GetByID(params GetPhotoParams) (map[string]any, error) {
	var photo model.Photo
	if err := model.DB.Preload("Author").First(&photo, params.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	result := map[string]any{
		"id":             photo.ID,
		"title":          photo.Title,
		"description":    photo.Description,
		"image_url":      photo.ImageURL,
		"author":         UserBrief{ID: photo.Author.ID, Name: photo.Author.Name},
		"solved":         photo.Solved,
		"attempts_count": photo.AttemptsCount,
		"created_at":     photo.CreatedAt,
	}

	// 如果已破解，返回获奖者信息
	if photo.Solved {
		var winnerAttempt model.Attempt
		if err := model.DB.Where("photo_id = ? AND is_winner = ?", params.PhotoID, true).
			Preload("User").First(&winnerAttempt).Error; err == nil {
			result["winner"] = map[string]any{
				"user_id":    winnerAttempt.UserID,
				"name":       winnerAttempt.User.Name,
				"created_at": winnerAttempt.ReviewedAt,
			}
		}
	}

	// 如果已登录，返回当前用户的答题记录
	if params.CurrentUserID > 0 {
		var userAttempt model.Attempt
		if err := model.DB.Where("photo_id = ? AND user_id = ?", params.PhotoID, params.CurrentUserID).
			First(&userAttempt).Error; err == nil {
			result["current_user_attempt"] = map[string]any{
				"id":        userAttempt.ID,
				"status":    userAttempt.Status,
				"is_winner": userAttempt.IsWinner,
			}
		}
	}

	return result, nil
}

// UserBrief 用户简要信息
type UserBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// saveUploadedFile 保存上传文件，返回原图URL和缩略图URL（当前版本二者相同）
func saveUploadedFile(file *multipart.FileHeader, subDir string) (string, string, error) {
	src, err := file.Open()
	if err != nil {
		return "", "", common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	// 校验文件类型
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", "", common.ErrNew(errors.New("图片必须为 jpg/png 格式"), common.ParamErr)
	}

	// 校验文件大小 (≤20MB)
	if file.Size > 20*1024*1024 {
		return "", "", common.ErrNew(errors.New("图片大小不能超过 20MB"), common.ParamErr)
	}

	// 生成文件名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	uploadDir := filepath.Join("uploads", subDir)
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", "", common.ErrNew(err, common.SysErr)
	}

	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", "", common.ErrNew(err, common.SysErr)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", "", common.ErrNew(err, common.SysErr)
	}

	url := fmt.Sprintf("/uploads/%s/%s", subDir, filename)
	return url, url, nil
}
