package service

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	_ "image/png"
	"math"
	"mime/multipart"
	"path/filepath"
	"strings"

	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/model"

	"github.com/disintegration/imaging"
	"gorm.io/gorm"
)

type PhotoSvc struct{}

// Create 上传图片投稿
func (info *PhotoSvc) Create(params PhotoCreateParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var activity model.Activity
	if err := tx.Where("id = ?", params.ActivityID).First(&activity).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("没有找到相应活动ID"), common.ParamErr)
	}

	var user model.User
	if err := tx.Where("id = ?", params.UserID).First(&user).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}
	// 保存图片
	imageURL, thumbURL, err := saveUploadedFile(params.ImageFile, "photos")
	if err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	gcjLat, gcjLng := WGS84orGCJ02ToGCJ02(params.Latitude, params.Longitude, params.CoordType)

	status := "pending"
	if config.Config.AUTO_APPROVAL == "all" {
		//自动审核
		distance1 := DistanceBetweenGCJ02(108.979167, 34.247222, gcjLat, gcjLng) //兴庆校区
		distance2 := DistanceBetweenGCJ02(108.941044, 34.216977, gcjLat, gcjLng) //雁塔校区
		distance3 := DistanceBetweenGCJ02(108.655162, 34.256229, gcjLat, gcjLng) //曲江校区
		distance4 := DistanceBetweenGCJ02(108.648747, 34.255606, gcjLat, gcjLng) //创新港校区
		if distance1 <= 1000 || distance2 <= 1000 || distance3 <= 1000 || distance4 <= 1000 {
			status = "approved"
		} else {
			status = "rejected"
		}
	}
	photo := &model.Photo{
		UserID:        params.UserID,
		ActivityID:    params.ActivityID,
		Title:         params.Title,
		Description:   params.Description,
		Latitude:      gcjLat,
		Longitude:     gcjLng,
		ImageURL:      imageURL,
		ThumbURL:      thumbURL,
		Status:        status,
		Solved:        false,
		AttemptsCount: 0,
		LikesCount:    0,
	}

	if err := tx.Create(photo).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	resp = ResponseIS{
		ID:     photo.ID,
		Status: photo.Status,
	}
	return resp, nil
}

// List 获取已审核通过的图片列表
func (info *PhotoSvc) List(params PhotoListParams) (resp PhotoForms, err error) {
	var photos []model.Photo
	var total int64
	query := model.DB.Model(&model.Photo{}).Where("status = ?", "approved")

	if params.Solved != nil {
		query = query.Where("solved = ?", *params.Solved)
	}
	if params.Keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+params.Keyword+"%", "%"+params.Keyword+"%")
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
			ID:         ph.ID,
			Title:      ph.Title,
			ThumbURL:   ph.ThumbURL,
			Author:     UserBrief{ID: ph.Author.ID, Nickname: ph.Author.Nickname, AvatarURL: ph.Author.AvatarURL},
			Solved:     ph.Solved,
			LikesCount: ph.LikesCount,
		})
	}

	resp = PhotoForms{
		Total:  total,
		Photos: photoForms,
	}
	return resp, nil
}

// GetByID 获取图片详情
func (info *PhotoSvc) GetByID(params PhotoGetByIDParams) (resp PhotoDetail, err error) {
	var photo model.Photo
	if err := model.DB.Preload("Author"). //预加载作者信息
						First(&photo, params.PhotoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = PhotoDetail{
		ID:            photo.ID,
		ActivityID:    photo.ActivityID,
		Title:         photo.Title,
		Description:   photo.Description,
		ImageURL:      photo.ImageURL,
		Author:        UserBrief{ID: photo.Author.ID, Nickname: photo.Author.Nickname, AvatarURL: photo.Author.AvatarURL},
		Solved:        photo.Solved,
		AttemptsCount: photo.AttemptsCount,
		LikesCount:    photo.LikesCount,
		CreatedAt:     photo.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

// GetImageStream 获取图片流（用于展示/下载），优先使用原图 URL
func (info *PhotoSvc) GetImageStream(photoID int64) (image ImageStream, err error) {
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

// ListByUser 获取某用户投稿的图片列表
func (info *PhotoSvc) ListByUser(params PhotosListUserParams) (resp PhotoForms, err error) {
	var photos []model.Photo
	var total int64
	query := model.DB.Model(&model.Photo{}).Where("user_id = ", params.UserID)

	if params.ActivityID > 0 {
		query = query.Where("activity_id = ?", params.ActivityID)
	}
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
			ID:         ph.ID,
			Title:      ph.Title,
			ThumbURL:   ph.ThumbURL,
			Author:     UserBrief{ID: ph.Author.ID, Nickname: ph.Author.Nickname, AvatarURL: ph.Author.AvatarURL},
			Solved:     ph.Solved,
			LikesCount: ph.LikesCount,
		})
	}

	resp = PhotoForms{
		Total:  total,
		Photos: photoForms,
	}
	return resp, nil
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
		return nil, common.ErrNew(err, common.SysErr)
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

// 坐标系转换
// 常量定义（这些常数是拟合参数，无需修改）
const (
	pi  = 3.14159265358979324
	a   = 6378245.0              // 长半轴
	ee  = 0.00669342162296594323 // 偏心率平方
	xPi = pi * 3000.0 / 180.0
)

// 判断坐标是否在中国境内（纬度 3.86~53.55，经度 73.66~135.05）
// 若不在中国境内，GCJ-02 和 WGS-84 相同，无需转换
func outOfChina(lat, lng float64) bool {
	return lng < 72.004 || lng > 137.8347 || lat < 0.8293 || lat > 55.8271
}

// 计算经度偏移（内部函数）
func transformLng(lat, lng float64) float64 {
	ret := 300.0 + lng + 2.0*lat + 0.1*lng*lng + 0.1*lng*lat + 0.1*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*pi) + 20.0*math.Sin(2.0*lng*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lng*pi) + 40.0*math.Sin(lng/3.0*pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(lng/12.0*pi) + 300.0*math.Sin(lng/30.0*pi)) * 2.0 / 3.0
	return ret
}

// 计算纬度偏移（内部函数）
func transformLat(lat, lng float64) float64 {
	ret := -100.0 + 2.0*lng + 3.0*lat + 0.2*lat*lat + 0.1*lng*lat + 0.2*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*pi) + 20.0*math.Sin(2.0*lng*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lat*pi) + 40.0*math.Sin(lat/3.0*pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(lat/12.0*pi) + 320.0*math.Sin(lat/30.0*pi)) * 2.0 / 3.0
	return ret
}

// WGS-84 转 GCJ-02
func WGS84ToGCJ02(wgsLat, wgsLng float64) (gcjLat, gcjLng float64) {
	if outOfChina(wgsLat, wgsLng) {
		return wgsLat, wgsLng
	}
	dLat := transformLat(wgsLat-35.0, wgsLng-105.0)
	dLng := transformLng(wgsLat-35.0, wgsLng-105.0)
	radLat := wgsLat / 180.0 * pi
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * pi)
	dLng = (dLng * 180.0) / (a / sqrtMagic * math.Cos(radLat) * pi)
	gcjLat = wgsLat + dLat
	gcjLng = wgsLng + dLng
	return
}

func WGS84orGCJ02ToGCJ02(wgsLat, wgsLng float64, CoordType string) (gcjLat, gcjLng float64) {
	if CoordType == "WGS84" {
		return WGS84ToGCJ02(wgsLat, wgsLng)
	}
	return wgsLat, wgsLng
}

// GCJ-02 转 WGS-84（迭代逼近法，通常 5 次迭代即可）
func GCJ02ToWGS84(gcjLat, gcjLng float64) (wgsLat, wgsLng float64) {
	if outOfChina(gcjLat, gcjLng) {
		return gcjLat, gcjLng
	}
	// 初始假设 WGS-84 = GCJ-02（迭代起点）
	wgsLat, wgsLng = gcjLat, gcjLng
	for i := 0; i < 5; i++ {
		// 当前估算的 WGS-84 转为 GCJ-02
		dLat, dLng := WGS84ToGCJ02(wgsLat, wgsLng)
		// 计算误差
		dLat = dLat - gcjLat
		dLng = dLng - gcjLng
		// 修正估算值
		wgsLat = wgsLat - dLat
		wgsLng = wgsLng - dLng
	}
	return
}
