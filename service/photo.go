package service

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"
	"path/filepath"
	"strings"

	"tu-xun/common"
	"tu-xun/model"

	"github.com/disintegration/imaging"
	"gorm.io/gorm"
)

type Photo struct{}

// Create 上传图片投稿
func (info *Photo) Create(params CreatePhotoParams) (resp CreatePhotoResponse, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 保存图片
	imageURL, thumbURL, err := saveUploadedFile(params.ImageFile, "photos")
	if err != nil {
		tx.Rollback()
		return resp, err
	}

	photo := &model.Photo{
		NetID:          params.NetID,
		Title:          params.Title,
		Description:    params.Description,
		ImageURL:       imageURL,
		ThumbURL:       thumbURL,
		LocationSecret: params.LocationSecret,
		Status:         "pending",
		Solved:         false,
		AttemptsCount:  0,
		LikesCount:     0,
	}

	if err := tx.Create(photo).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	resp = CreatePhotoResponse{
		ID:      photo.ID,
		Message: "图片上传成功，正在审核中",
	}
	return resp, nil
}

// List 获取已审核通过的图片列表
func (info *Photo) List(params ListPhotoParams) (resp ListPhotosResponse, err error) {
	var photos []model.Photo
	var total int64
	query := model.DB.Model(&model.Photo{}).Where("status = ?", "approved")
	if params.Solved != nil {
		query = query.Where("solved = ?", *params.Solved)
	}
	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	switch params.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "attempts_count":
		query = query.Order("attempts_count DESC")
	case "likes_count":
		query = query.Order("likes_count DESC")
	default:
		query = query.Order("created_at DESC")
	}
	if err := query.Preload("Author").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&photos).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 隐藏敏感字段
	photoForms := make([]PhotoForm, 0, len(photos))
	for _, ph := range photos {
		photoForms = append(photoForms, PhotoForm{
			ID:            ph.ID,
			Title:         ph.Title,
			Description:   ph.Description,
			ThumbURL:      ph.ThumbURL,
			Author:        UserBrief{ID: ph.Author.ID, Name: ph.Author.Name, AvatarURL: ph.Author.AvatarURL},
			Solved:        ph.Solved,
			CreatedAt:     ph.CreatedAt,
			AttemptsCount: ph.AttemptsCount,
			LikesCount:    ph.LikesCount,
		})
	}

	resp = ListPhotosResponse{
		Total:  total,
		Photos: photoForms,
	}
	return resp, nil
}

// GetByID 获取图片详情
func (info *Photo) GetByID(params GetPhotoParams) (resp PhotoDetailResponse, err error) {
	var photo model.Photo
	if err := model.DB.Preload("Author"). //预加载作者信息
						First(&photo, params.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = PhotoDetailResponse{
		ID:            photo.ID,
		Title:         photo.Title,
		Description:   photo.Description,
		ImageURL:      photo.ImageURL,
		Author:        UserBrief{ID: photo.Author.ID, Name: photo.Author.Name, AvatarURL: photo.Author.AvatarURL},
		Solved:        photo.Solved,
		AttemptsCount: photo.AttemptsCount,
		CreatedAt:     photo.CreatedAt,
		Winner:        AttemptForm{}, // 默认无获奖者信息，后续如果已破解会补充
	}

	// 如果已破解，返回获奖者信息
	if photo.Solved {
		var winnerAttempt model.Attempt
		if err := model.DB.Where("photo_id = ? AND solved = ?", params.PhotoID, 2).
			Preload("User").First(&winnerAttempt).Error; err == nil {
			resp.Winner = AttemptForm{
				ID:              winnerAttempt.ID,
				ImageURL:        winnerAttempt.ImageURL,
				GuessedLocation: winnerAttempt.GuessedLocation,
				CreatedAt:       winnerAttempt.CreatedAt,
				User:            UserBrief{ID: winnerAttempt.User.ID, Name: winnerAttempt.User.Name, AvatarURL: winnerAttempt.User.AvatarURL},
			}
		}
	}
	return resp, nil
}

// GetImageStream 获取图片流（用于展示/下载），优先使用原图 URL
func (info *Photo) GetImageStream(photoID int64) (image ImageStream, err error) {
	var photo model.Photo
	if err := model.DB.First(&photo, photoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return image, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return image, common.ErrNew(err, common.SysErr)
	}

	if photo.ImageURL == "" {
		return image, common.ErrNew(errors.New("图片 URL 为空"), common.SysErr)
	}

	// 从 OSS URL 提取 object key
	objectKey := OSSClient.ExtractObjectKey(photo.ImageURL)

	// 从 OSS 获取文件流
	reader, contentType, size, err := OSSClient.GetObject(objectKey)
	if err != nil {
		return image, common.ErrNew(err, common.SysErr)
	}

	// 从 URL 提取文件名
	filename := filepath.Base(objectKey)

	return ImageStream{
		Reader:      reader,
		ContentType: contentType,
		Size:        size,
		Filename:    filename,
	}, nil
}

// saveUploadedFile 保存上传文件，同时生成缩略图，返回原图URL和缩略图URL
func saveUploadedFile(file *multipart.FileHeader, subDir string) (string, string, error) {
	src, err := file.Open()
	if err != nil {
		return "", "", common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", "", common.ErrNew(errors.New("图片必须为 jpg/png 格式"), common.ParamErr)
	}

	// 校验文件大小 (≤20MB)
	if file.Size > 20*1024*1024 {
		return "", "", common.ErrNew(errors.New("图片大小不能超过 20MB"), common.ParamErr)
	}

	// 解码原图
	img, _, err := image.Decode(src)
	if err != nil {
		return "", "", common.ErrNew(errors.New("无法解码图片，请确认文件为有效图片"), common.ParamErr)
	}

	// 生成缩略图（最大宽度 400px，JPEG 质量 80%）
	thumbData, err := generateThumbnail(img, ext)
	if err != nil {
		return "", "", common.ErrNew(err, common.SysErr)
	}

	// 上传原图
	imageURL, err := OSSClient.UploadFile(file, subDir)
	if err != nil {
		return "", "", common.ErrNew(err, common.SysErr)
	}

	// 上传缩略图（缩略图统一用 .jpg 格式）
	thumbFilename := strings.TrimSuffix(file.Filename, ext) + "_thumb.jpg"
	thumbURL, err := OSSClient.UploadBytes(thumbData, thumbFilename, subDir)
	if err != nil {
		return "", "", common.ErrNew(err, common.SysErr)
	}

	return imageURL, thumbURL, nil
}

// generateThumbnail 生成缩略图字节数据（最大宽度 400px，JPEG 格式）
func generateThumbnail(img image.Image, originalExt string) ([]byte, error) {
	// 缩略图最大宽度
	const maxWidth = 400
	if originalExt != ".jpg" && originalExt != ".jpeg" && originalExt != ".png" {
		return nil, common.ErrNew(errors.New("图片必须为 jpg/png 格式"), common.ParamErr)
	}
	bounds := img.Bounds()
	w := bounds.Dx()
	// h := bounds.Dy()

	// 如果原图宽度小于缩略图最大宽度，不缩小
	var thumbnail image.Image
	if w <= maxWidth {
		thumbnail = img
	} else {
		thumbnail = imaging.Resize(img, maxWidth, 0, imaging.Lanczos)
	}

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, thumbnail, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// saveThumbnailOnly 仅生成并保存缩略图（用于答题图片等只需缩略图的场景）
func saveThumbnailOnly(file *multipart.FileHeader, subDir string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", common.ErrNew(errors.New("图片必须为 jpg/png 格式"), common.ParamErr)
	}

	// 校验文件大小 (≤20MB)
	if file.Size > 20*1024*1024 {
		return "", common.ErrNew(errors.New("图片大小不能超过 20MB"), common.ParamErr)
	}

	// 解码原图
	img, _, err := image.Decode(src)
	if err != nil {
		return "", common.ErrNew(errors.New("无法解码图片，请确认文件为有效图片"), common.ParamErr)
	}

	// 生成缩略图
	thumbData, err := generateThumbnail(img, ext)
	if err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}

	// 只上传缩略图
	thumbFilename := strings.TrimSuffix(file.Filename, ext) + "_thumb.jpg"
	thumbURL, err := OSSClient.UploadBytes(thumbData, thumbFilename, subDir)
	if err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}

	return thumbURL, nil
}

// ListByUser 获取某用户投稿的图片列表
func (info *Photo) ListByUser(params ListUserPhotosParams) (resp ListPhotosResponse, err error) {
	var photos []model.Photo
	var total int64

	query := model.DB.Model(&model.Photo{}).Where("user_id = ? AND status = ?", params.NetID, "approved")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	switch params.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "attempts_count":
		query = query.Order("attempts_count DESC")
	case "likes_count":
		query = query.Order("likes_count DESC")
	default:
		query = query.Order("created_at DESC")
	}
	if err := query.Preload("Author").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&photos).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	photoForms := make([]PhotoForm, 0, len(photos))
	for _, ph := range photos {
		photoForms = append(photoForms, PhotoForm{
			ID:            ph.ID,
			Title:         ph.Title,
			Description:   ph.Description,
			ThumbURL:      ph.ThumbURL,
			Author:        UserBrief{ID: ph.Author.ID, Name: ph.Author.Name, AvatarURL: ph.Author.AvatarURL},
			Solved:        ph.Solved,
			CreatedAt:     ph.CreatedAt,
			AttemptsCount: ph.AttemptsCount,
			LikesCount:    ph.LikesCount,
		})
	}

	resp = ListPhotosResponse{
		Total:  total,
		Photos: photoForms,
	}
	return resp, nil
}
