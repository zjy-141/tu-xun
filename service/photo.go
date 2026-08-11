package service

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"math"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	// 验证活动未结束
	now := time.Now()
	if activity.EndTime != nil && !now.Before(*activity.EndTime) || activity.StartTime != nil && now.Before(*activity.StartTime) {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("不在活动时间内，无法投稿"), common.ParamErr)
	}

	// 保存图片
	uploadResult, err := saveUploadedFile(params.ImageFile, "photos", true)
	if err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	gcjLat, gcjLng := WGS84orGCJ02ToGCJ02(params.Latitude, params.Longitude, params.CoordType)

	status := "pending"
	rejectReason := ""
	if config.Config.AUTO_APPROVAL == "all" {
		// 自动审核：校园范围检查
		distance1 := DistanceBetweenGCJ02(108.979167, 34.247222, gcjLng, gcjLat) // 兴庆校区
		distance2 := DistanceBetweenGCJ02(108.941044, 34.216977, gcjLng, gcjLat) // 雁塔校区
		distance3 := DistanceBetweenGCJ02(108.655162, 34.256229, gcjLng, gcjLat) // 曲江校区
		distance4 := DistanceBetweenGCJ02(108.648747, 34.255606, gcjLng, gcjLat) // 创新港校区
		if distance1 <= 1000 || distance2 <= 1000 || distance3 <= 1000 || distance4 <= 1000 {
			status = "approved"
		} else {
			status = "rejected"
			rejectReason = "投稿位置不在校园范围内，无法通过审核"
		}
	}

	photo := &model.Photo{
		UserID:        params.UserID,
		ActivityID:    params.ActivityID,
		Title:         params.Title,
		Description:   params.Description,
		Latitude:      gcjLat,
		Longitude:     gcjLng,
		CoordType:     "gcj02",
		ImageURL:      uploadResult.ImageURL,
		ThumbURL:      uploadResult.ThumbURL,
		ImageWidth:    uploadResult.ImageWidth,
		ImageHeight:   uploadResult.ImageHeight,
		ThumbWidth:    uploadResult.ThumbWidth,
		ThumbHeight:   uploadResult.ThumbHeight,
		Status:        status,
		RejectReason:  rejectReason,
		Solved:        false,
		SolvedCount:   0,
		AttemptsCount: 0,
		LikesCount:    0,
	}

	if err := tx.Create(photo).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 自动审核：在事务内完成积分发放
	if config.Config.AUTO_APPROVAL == "all" {
		now := time.Now()
		photo.Status = status
		photo.ReviewedAt = &now

		if status == "approved" {
			scoreSvc := ScoreSvc{}
			if _, err := scoreSvc.RegularScoreChange(tx, ScoreChangeParams{
				UserID:      photo.UserID,
				Delta:       activity.PhotoPoints,
				Reason:      "upload_photo",
				RelatedID:   photo.ID,
				RelatedType: "photo",
			}); err != nil {
				tx.Rollback()
				return resp, common.ErrNew(err, common.SysErr)
			}
		}

		// 持久化审核状态和审核时间
		if err := tx.Save(&photo).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}
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

// List 获取已审核通过的题目列表
func (info *PhotoSvc) List(params PhotoListParams, userID int64) (resp PhotoCardPage, err error) {
	var photos []model.Photo
	var total int64
	query := model.DB.Model(&model.Photo{}).Where("status = ?", "approved")
	if params.ActivityID > 0 {
		query = query.Where("activity_id = ?", params.ActivityID)
	}
	if params.Solved != nil {
		if *params.Solved {
			query = query.Where("solved_count > 0")
		} else {
			query = query.Where("solved_count = 0")
		}
	}
	if params.Keyword != "" {
		like := "%" + params.Keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	// 解析热度权重（默认 likes×2 + attempts×1）
	likeW := parseIntOr(config.Config.HOT_LIKE_WEIGHT, 2)
	attemptW := parseIntOr(config.Config.HOT_ATTEMPT_WEIGHT, 1)

	switch params.SortBy {
	case "hot":
		query = query.Order(fmt.Sprintf("(likes_count * %d + attempts_count * %d) DESC, id DESC", likeW, attemptW))
	case "created_at":
		query = query.Order("created_at DESC, id DESC")
	default:
		query = query.Order("created_at DESC, id DESC")
	}
	if err := query.Preload("Author").Preload("Activity").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&photos).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 批量查询当前用户的点赞状态
	likedSet := make(map[int64]bool)
	if userID > 0 && len(photos) > 0 {
		photoIDs := make([]int64, len(photos))
		for i, ph := range photos {
			photoIDs[i] = ph.ID
		}
		var likes []model.Like
		model.DB.Where("user_id = ? AND target_type = ? AND target_id IN ?", userID, "photo", photoIDs).
			Find(&likes)
		for _, l := range likes {
			likedSet[l.TargetID] = true
		}
	}

	// 批量查询当前用户的破解状态
	solvedSet := make(map[int64]bool)
	if userID > 0 && len(photos) > 0 {
		photoIDs := make([]int64, len(photos))
		for i, ph := range photos {
			photoIDs[i] = ph.ID
		}
		var attempts []model.Attempt
		model.DB.Select("photo_id").Where("user_id = ? AND photo_id IN ? AND status = ?", userID, photoIDs, "solved").
			Find(&attempts)
		for _, a := range attempts {
			solvedSet[a.PhotoID] = true
		}
	}

	cards := make([]PhotoCard, 0, len(photos))
	for _, ph := range photos {
		cards = append(cards, PhotoCard{
			ID:       ph.ID,
			Activity: ActivityBrief{ID: ph.Activity.ID, Title: ph.Activity.Title, StartTime: ph.Activity.StartTime, EndTime: ph.Activity.EndTime},
			Author:   UserBrief{ID: ph.Author.ID, Nickname: ph.Author.Nickname, Avatar: ph.Author.AvatarURL},
			Title:    ph.Title,
			Image: Media{
				ThumbURL: ph.ThumbURL,
				Width:    ph.ThumbWidth,
				Height:   ph.ThumbHeight,
			},
			LikesCount: ph.LikesCount,
			Liked:      likedSet[ph.ID],
			Solved:     solvedSet[ph.ID],
			CreatedAt:  &ph.CreatedAt,
		})
	}

	resp = PhotoCardPage{
		Total: total,
		List:  cards,
	}
	return resp, nil
}

// GetByID 获取题目详情
func (info *PhotoSvc) GetByID(photoID int64, userID int64) (*PhotoDetail, error) {
	var photo model.Photo
	if err := model.DB.Preload("Author").Preload("Activity").
		First(&photo, photoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	resp := &PhotoDetail{
		ID:          photo.ID,
		Activity:    ActivityBrief{ID: photo.Activity.ID, Title: photo.Activity.Title, StartTime: photo.Activity.StartTime, EndTime: photo.Activity.EndTime},
		Author:      UserBrief{ID: photo.Author.ID, Nickname: photo.Author.Nickname, Avatar: photo.Author.AvatarURL},
		Title:       photo.Title,
		Description: photo.Description,
		Image: Media{
			OriginURL: photo.ImageURL,
			Width:     photo.ImageWidth,
			Height:    photo.ImageHeight,
		},
		Location:      nil,
		SolvedCount:   photo.SolvedCount,
		AttemptsCount: photo.AttemptsCount,
		LikesCount:    photo.LikesCount,
		CreatedAt:     &photo.CreatedAt,
		Status:        photo.Status,
	}

	// 活动已结束或当前用户是作者 → 返回坐标
	if photo.Activity.EndTime != nil && !time.Now().Before(*photo.Activity.EndTime) ||
		(userID > 0 && photo.UserID == userID) {
		resp.Location = &Location{
			Longitude: photo.Longitude,
			Latitude:  photo.Latitude,
			CoordType: photo.CoordType,
		}
	}

	// 查询当前用户是否已破解、答题次数、点赞状态
	if userID > 0 {
		// 用户答题次数
		var userAttemptsCount int64
		model.DB.Model(&model.Attempt{}).
			Where("photo_id = ? AND user_id = ?", photoID, userID).
			Count(&userAttemptsCount)
		resp.UserAttemptsCount = int(userAttemptsCount)

		// 用户是否已破解（有 solved 状态的答题记录）
		var solvedCount int64
		model.DB.Model(&model.Attempt{}).
			Where("photo_id = ? AND user_id = ? AND status = ?", photoID, userID, "solved").
			Count(&solvedCount)
		resp.Solved = solvedCount > 0

		// 用户是否已点赞
		var likeCount int64
		model.DB.Model(&model.Like{}).
			Where("user_id = ? AND target_type = ? AND target_id = ?", userID, "photo", photoID).
			Count(&likeCount)
		resp.Liked = likeCount > 0
	}

	return resp, nil
}

// ListUser 获取当前用户投稿的题目列表
func (info *PhotoSvc) ListUser(params PhotosListUserParams) (resp UserPhotoCardPage, err error) {
	var photos []model.Photo
	var total int64
	query := model.DB.Model(&model.Photo{}).Where("user_id = ?", params.UserID)

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	query = query.Order("created_at DESC")
	if err := query.Preload("Activity").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&photos).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	cards := make([]UserPhotoCard, 0, len(photos))
	for _, ph := range photos {
		cards = append(cards, UserPhotoCard{
			ID:       ph.ID,
			Activity: ActivityBrief{ID: ph.Activity.ID, Title: ph.Activity.Title, StartTime: ph.Activity.StartTime, EndTime: ph.Activity.EndTime},
			Title:    ph.Title,
			Image: Media{
				ThumbURL: ph.ThumbURL,
				Width:    ph.ThumbWidth,
				Height:   ph.ThumbHeight,
			},
			AttemptsCount: ph.AttemptsCount,
			SolvedCount:   ph.SolvedCount,
			Status:        ph.Status,
			CreatedAt:     &ph.CreatedAt,
		})
	}

	resp = UserPhotoCardPage{
		Total: total,
		List:  cards,
	}
	return resp, nil
}

// DetailUser 获取我的投稿详情（仅 pending/rejected 状态可查看）
func (info *PhotoSvc) DetailUser(photoID int64, userID int64) (*UserPhotoDetail, error) {
	var photo model.Photo
	if err := model.DB.Preload("Activity").
		First(&photo, photoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	// 仅允许查看本人的 pending/rejected 投稿
	if photo.UserID != userID || photo.Status == "approved" {
		return nil, common.ErrNew(errors.New("图片不存在"), common.OpErr)
	}

	resp := &UserPhotoDetail{
		ID:          photo.ID,
		Activity:    ActivityBrief{ID: photo.Activity.ID, Title: photo.Activity.Title, StartTime: photo.Activity.StartTime, EndTime: photo.Activity.EndTime},
		Title:       photo.Title,
		Description: photo.Description,
		Image: Media{
			OriginURL: photo.ImageURL,
			Width:     photo.ImageWidth,
			Height:    photo.ImageHeight,
		},
		Location: Location{
			Longitude: photo.Longitude,
			Latitude:  photo.Latitude,
			CoordType: photo.CoordType,
		},
		Status:       photo.Status,
		RejectReason: photo.RejectReason,
		CreatedAt:    &photo.CreatedAt,
	}

	return resp, nil
}

// ListSolves 获取题目的破解记录列表（仅 solved）
func (info *PhotoSvc) ListSolves(params SolvesListParams, userID int64) (resp SolveItemPage, err error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{}).
		Where("photo_id = ? AND status = ?", params.PhotoID, "solved")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("User").
		Order("created_at DESC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 批量查询当前用户的点赞状态
	likedSet := make(map[int64]bool)
	if userID > 0 && len(attempts) > 0 {
		attemptIDs := make([]int64, len(attempts))
		for i, a := range attempts {
			attemptIDs[i] = a.ID
		}
		var likes []model.Like
		model.DB.Where("user_id = ? AND target_type = ? AND target_id IN ?", userID, "attempt", attemptIDs).
			Find(&likes)
		for _, l := range likes {
			likedSet[l.TargetID] = true
		}
	}

	items := make([]SolveItem, 0, len(attempts))
	for _, a := range attempts {
		items = append(items, SolveItem{
			ID:     a.ID,
			Author: UserBrief{ID: a.User.ID, Nickname: a.User.Nickname, Avatar: a.User.AvatarURL},
			Image: Media{
				ThumbURL: a.ImageURL,
				Width:    a.ImageWidth,
				Height:   a.ImageHeight,
			},
			Location: Location{
				Longitude: a.Longitude,
				Latitude:  a.Latitude,
				CoordType: a.CoordType,
			},
			LikesCount: a.LikesCount,
			Liked:      likedSet[a.ID],
			CreatedAt:  &a.CreatedAt,
		})
	}

	resp = SolveItemPage{
		Total: total,
		List:  items,
	}
	return resp, nil
}

// PhotoComments 获取题目的评论列表（仅 approved）
func (info *PhotoSvc) PhotoComments(params CommentListParams, userID int64) (resp CommentItemPage, err error) {
	var comments []model.Comment
	var total int64

	query := model.DB.Model(&model.Comment{}).
		Where("photo_id = ? AND status = ?", params.PhotoID, "approved")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	switch params.SortBy {
	case "likes_count":
		query = query.Order("likes_count DESC, created_at DESC")
	case "created_at":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	if err := query.Preload("User").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&comments).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 批量查询当前用户的点赞状态
	likedSet := make(map[int64]bool)
	if userID > 0 && len(comments) > 0 {
		commentIDs := make([]int64, len(comments))
		for i, c := range comments {
			commentIDs[i] = c.ID
		}
		var likes []model.Like
		model.DB.Where("user_id = ? AND target_type = ? AND target_id IN ?", userID, "comment", commentIDs).
			Find(&likes)
		for _, l := range likes {
			likedSet[l.TargetID] = true
		}
	}

	items := make([]CommentItem, 0, len(comments))
	for _, c := range comments {
		items = append(items, CommentItem{
			ID:         c.ID,
			Author:     UserBrief{ID: c.User.ID, Nickname: c.User.Nickname, Avatar: c.User.AvatarURL},
			Content:    c.CommentText,
			LikesCount: c.LikesCount,
			Liked:      likedSet[c.ID],
			CreatedAt:  &c.CreatedAt,
		})
	}

	resp = CommentItemPage{
		Total: total,
		List:  items,
	}
	return resp, nil
}

// PhotoAttemptsUser 获取当前用户在某题目下的作答记录
func (info *PhotoSvc) PhotoAttemptsUser(params PhotoAttemptsUserListParams) (resp AttemptRecordPage, err error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{}).
		Where("photo_id = ? AND user_id = ?", params.PhotoID, params.UserID)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Order("created_at DESC").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	records := make([]AttemptRecord, 0, len(attempts))
	for _, a := range attempts {
		records = append(records, AttemptRecord{
			ID: a.ID,
			Image: Media{
				ThumbURL: a.ImageURL,
				Width:    a.ImageWidth,
				Height:   a.ImageHeight,
			},
			Location: Location{
				Longitude: a.Longitude,
				Latitude:  a.Latitude,
				CoordType: a.CoordType,
			},
			Status:       a.Status,
			RejectReason: a.RejectReason,
			CreatedAt:    &a.CreatedAt,
		})
	}

	resp = AttemptRecordPage{
		Total: total,
		List:  records,
	}
	return resp, nil
}

// ==================== 图片保存工具函数 ====================

// UploadResult 图片上传结果
type UploadResult struct {
	ImageURL    string
	ThumbURL    string
	ImageWidth  int
	ImageHeight int
	ThumbWidth  int
	ThumbHeight int
}

// saveUploadedFile 保存上传文件，同时生成缩略图，返回上传结果（含尺寸）
func saveUploadedFile(file *multipart.FileHeader, subDir string, thumb bool) (UploadResult, error) {
	src, err := file.Open()
	if err != nil {
		return UploadResult{}, common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return UploadResult{}, common.ErrNew(errors.New("图片必须为 jpg/png 格式"), common.ParamErr)
	}

	// 校验文件大小 (≤20MB)
	if file.Size > 20*1024*1024 {
		return UploadResult{}, common.ErrNew(errors.New("图片大小不能超过 20MB"), common.ParamErr)
	}

	// 解码原图
	img, _, err := image.Decode(src)
	if err != nil {
		return UploadResult{}, common.ErrNew(errors.New("无法解码图片，请确认文件为有效图片"), common.ParamErr)
	}

	bounds := img.Bounds()
	result := UploadResult{ImageWidth: bounds.Dx(), ImageHeight: bounds.Dy()}

	// 上传原图
	result.ImageURL, err = OSSClient.UploadFile(file, subDir)
	if err != nil {
		return UploadResult{}, common.ErrNew(err, common.SysErr)
	}
	// 生成缩略图（最大宽度 400px，JPEG 质量 80%）
	if thumb {
		thumbData, err := generateThumbnail(img, ext)
		if err != nil {
			return UploadResult{}, common.ErrNew(err, common.SysErr)
		}
		// 计算缩略图尺寸
		const maxWidth = 400
		if result.ImageWidth <= maxWidth {
			result.ThumbWidth = result.ImageWidth
			result.ThumbHeight = result.ImageHeight
		} else {
			result.ThumbWidth = maxWidth
			result.ThumbHeight = int(float64(result.ImageHeight) * float64(maxWidth) / float64(result.ImageWidth))
		}
		// 上传缩略图（缩略图统一用 .jpg 格式）
		thumbFilename := strings.TrimSuffix(file.Filename, ext) + "_thumb.jpg"
		result.ThumbURL, err = OSSClient.UploadBytes(thumbData, thumbFilename, subDir)
		if err != nil {
			return UploadResult{}, common.ErrNew(err, common.SysErr)
		}
	}
	return result, nil
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

// saveThumbnailOnly 仅生成并保存缩略图（用于答题图片等只需缩略图的场景），返回缩略图URL和尺寸
func saveThumbnailOnly(file *multipart.FileHeader, subDir string) (thumbURL string, thumbWidth int, thumbHeight int, err error) {
	src, err := file.Open()
	if err != nil {
		return "", 0, 0, common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", 0, 0, common.ErrNew(errors.New("图片必须为 jpg/png 格式"), common.ParamErr)
	}

	// 校验文件大小 (≤20MB)
	if file.Size > 20*1024*1024 {
		return "", 0, 0, common.ErrNew(errors.New("图片大小不能超过 20MB"), common.ParamErr)
	}

	// 解码原图
	img, _, err := image.Decode(src)
	if err != nil {
		return "", 0, 0, common.ErrNew(errors.New("无法解码图片，请确认文件为有效图片"), common.ParamErr)
	}

	bounds := img.Bounds()
	origW, origH := bounds.Dx(), bounds.Dy()

	// 生成缩略图
	thumbData, err := generateThumbnail(img, ext)
	if err != nil {
		return "", 0, 0, common.ErrNew(err, common.SysErr)
	}

	// 计算缩略图尺寸
	const maxWidth = 400
	if origW <= maxWidth {
		thumbWidth = origW
		thumbHeight = origH
	} else {
		thumbWidth = maxWidth
		thumbHeight = int(float64(origH) * float64(maxWidth) / float64(origW))
	}

	// 只上传缩略图
	thumbFilename := strings.TrimSuffix(file.Filename, ext) + "_thumb.jpg"
	thumbURL, err = OSSClient.UploadBytes(thumbData, thumbFilename, subDir)
	if err != nil {
		return "", 0, 0, common.ErrNew(err, common.SysErr)
	}

	return thumbURL, thumbWidth, thumbHeight, nil
}

// saveUploadedMedia 保存反馈附件（支持图片和视频），返回上传结果
func saveUploadedMedia(file *multipart.FileHeader, subDir string) (originURL string, thumbURL string, width int, height int, mediaType int, err error) {
	src, err := file.Open()
	if err != nil {
		return "", "", 0, 0, 0, common.ErrNew(err, common.SysErr)
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	contentType := file.Header.Get("Content-Type")

	// 判断是否视频
	isVideo := strings.HasPrefix(contentType, "video/") || ext == ".mp4" || ext == ".mov" || ext == ".avi" || ext == ".webm"

	if isVideo {
		// 视频：最大 50MB
		if file.Size > 50*1024*1024 {
			return "", "", 0, 0, 0, common.ErrNew(errors.New("视频大小不能超过 50MB"), common.ParamErr)
		}
		originURL, err = OSSClient.UploadFile(file, subDir)
		if err != nil {
			return "", "", 0, 0, 0, common.ErrNew(err, common.SysErr)
		}
		return originURL, "", 0, 0, 2, nil
	}

	// 图片：走原有逻辑
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", "", 0, 0, 0, common.ErrNew(errors.New("图片必须为 jpg/png 格式"), common.ParamErr)
	}
	if file.Size > 20*1024*1024 {
		return "", "", 0, 0, 0, common.ErrNew(errors.New("图片大小不能超过 20MB"), common.ParamErr)
	}

	img, _, err := image.Decode(src)
	if err != nil {
		return "", "", 0, 0, 0, common.ErrNew(errors.New("无法解码图片，请确认文件为有效图片"), common.ParamErr)
	}

	bounds := img.Bounds()
	width, height = bounds.Dx(), bounds.Dy()

	// 上传原图
	originURL, err = OSSClient.UploadFile(file, subDir)
	if err != nil {
		return "", "", 0, 0, 0, common.ErrNew(err, common.SysErr)
	}

	// 生成并上传缩略图
	thumbData, err := generateThumbnail(img, ext)
	if err != nil {
		return "", "", 0, 0, 0, common.ErrNew(err, common.SysErr)
	}
	thumbFilename := strings.TrimSuffix(file.Filename, ext) + "_thumb.jpg"
	thumbURL, err = OSSClient.UploadBytes(thumbData, thumbFilename, subDir)
	if err != nil {
		return "", "", 0, 0, 0, common.ErrNew(err, common.SysErr)
	}

	return originURL, thumbURL, width, height, 1, nil
}

// ==================== 坐标系转换 ====================

// 常量定义（这些常数是拟合参数，无需修改）
const (
	pi  = 3.14159265358979324
	a   = 6378245.0              // 长半轴
	ee  = 0.00669342162296594323 // 偏心率平方
	xPi = pi * 3000.0 / 180.0
)

// 判断坐标是否在中国境内（纬度 3.86~53.55，经度 73.66~135.05）
func outOfChina(lat, lng float64) bool {
	return lng < 72.004 || lng > 137.8347 || lat < 0.8293 || lat > 55.8271
}

func transformLng(lat, lng float64) float64 {
	ret := 300.0 + lng + 2.0*lat + 0.1*lng*lng + 0.1*lng*lat + 0.1*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*pi) + 20.0*math.Sin(2.0*lng*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lng*pi) + 40.0*math.Sin(lng/3.0*pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(lng/12.0*pi) + 300.0*math.Sin(lng/30.0*pi)) * 2.0 / 3.0
	return ret
}

func transformLat(lat, lng float64) float64 {
	ret := -100.0 + 2.0*lng + 3.0*lat + 0.2*lat*lat + 0.1*lng*lat + 0.2*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*pi) + 20.0*math.Sin(2.0*lng*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lat*pi) + 40.0*math.Sin(lat/3.0*pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(lat/12.0*pi) + 320.0*math.Sin(lat/30.0*pi)) * 2.0 / 3.0
	return ret
}

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

func GCJ02ToWGS84(gcjLat, gcjLng float64) (wgsLat, wgsLng float64) {
	if outOfChina(gcjLat, gcjLng) {
		return gcjLat, gcjLng
	}
	wgsLat, wgsLng = gcjLat, gcjLng
	for i := 0; i < 5; i++ {
		dLat, dLng := WGS84ToGCJ02(wgsLat, wgsLng)
		dLat = dLat - gcjLat
		dLng = dLng - gcjLng
		wgsLat = wgsLat - dLat
		wgsLng = wgsLng - dLng
	}
	return
}

// parseIntOr 解析字符串为 int，解析失败时返回默认值
func parseIntOr(s string, defaultVal int) int {
	if v, err := strconv.Atoi(s); err == nil && v >= 0 {
		return v
	}
	return defaultVal
}
