package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/model"

	"gorm.io/gorm"
)

type AttemptSvc struct{}

// Create 提交答题
func (a *AttemptSvc) Create(info AttemptCreateParams) (resp ResponseIS, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 检查图片是否存在且已审核通过
	var photo model.Photo
	if err := tx.Preload("Activity.AttemptRewardTiers").First(&photo, info.PhotoID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}
	if photo.Status != "approved" {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该图片尚未通过审核，暂不可答题"), common.OpErr)
	}

	// 检查活动是否进行中
	activitySvc := ActivitySvc{}
	if active, _ := activitySvc.IsActivityActive(photo.ActivityID); !active {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该活动不在进行中"), common.ParamErr)
	}

	// 检查是否已有过多的答题记录（同一用户同一图片）
	var existtotal int64
	if err := tx.Where("photo_id = ? AND user_id = ?", info.PhotoID, info.UserID).
		Count(&existtotal).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}
	if existtotal > 100 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("您有过多的答题记录待审核，请耐心等待管理员审核"), common.ParamErr)
	}

	// 保存答题图片（仅缩略图）
	imageURL, err := saveThumbnailOnly(info.ImageFile, "attempts")
	if err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	gcjLat, gcjLng := WGS84orGCJ02ToGCJ02(info.Latitude, info.Longitude, info.CoordType)

	status := "pending"
	if config.Config.AUTO_APPROVAL == "attemptAndComment" || config.Config.AUTO_APPROVAL == "all" {
		//自动审核
		distance := DistanceBetweenGCJ02(photo.Latitude, photo.Longitude, gcjLat, gcjLng)
		if distance <= 50 {
			status = "solved"
		} else {
			status = "unsolved"
		}
	}

	attempt := &model.Attempt{
		PhotoID:     info.PhotoID,
		UserID:      info.UserID,
		CommentText: info.AnswerText,
		ImageURL:    imageURL,
		Latitude:    gcjLat,
		Longitude:   gcjLng,
		LikesCount:  0,
		Status:      status,
	}

	if err := tx.Create(attempt).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 自动审核：在事务内完成积分发放、通知和计数更新
	if config.Config.AUTO_APPROVAL == "attemptAndComment" || config.Config.AUTO_APPROVAL == "all" {
		now := time.Now()
		attempt.Status = status
		attempt.ReviewedAt = &now

		if status == "solved" {
			// 只有答对时才发放积分和标记图片已破解
			if attempt.UserID != photo.UserID {
				rank, err := activitySvc.GetUserRank(attempt.UserID, attempt.PhotoID)
				if err != nil {
					tx.Rollback()
					return resp, common.ErrNew(err, common.SysErr)
				}

				var delta, awardedBatch int
				tiers := photo.Activity.AttemptRewardTiers
				sort.Slice(tiers, func(i, j int) bool {
					return tiers[i].Batch < tiers[j].Batch
				})
				if rank > 0 {
					for _, tier := range tiers {
						if rank <= tier.RankLimit {
							delta = tier.AttemptPoints
							awardedBatch = tier.Batch
							break
						}
					}
				}

				scoreSvc := ScoreSvc{}
				if _, err := scoreSvc.RegularScoreChange(tx, ScoreChangeParams{
					UserID:      attempt.UserID,
					Delta:       delta,
					Reason:      "upload_photo",
					RelatedID:   attempt.ID,
					RelatedType: "attempt",
					Remark:      fmt.Sprintf("恭喜你答对了，是第 %d 批次，得分 %d ！", awardedBatch, delta),
				}); err != nil {
					tx.Rollback()
					return resp, common.ErrNew(err, common.SysErr)
				}
			}

			if err := tx.Model(&model.Photo{}).Where("id = ?", attempt.PhotoID).Update("solved", true).Error; err != nil {
				tx.Rollback()
				return resp, common.ErrNew(err, common.SysErr)
			}
		}

		// 更新答题次数
		if err := tx.Model(&model.Photo{}).
			Where("id = ?", attempt.PhotoID).
			Update("attempts_count", gorm.Expr("attempts_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}

		// 发送通知
		msgTitle := "您的答题正确"
		msgType := "review"
		msgCategory := "normal"
		msgContent := "恭喜你答对了！"
		if status != "solved" {
			msgTitle = "您的答题不正确"
			msgType = "review"
			msgCategory = "normal"
			msgContent = "您的答题不正确，未能获得奖品。别气馁，失败是常态，调整心态再试试吧！"
		}
		msg := &model.Message{
			UserID:      attempt.UserID,
			SenderID:    1,
			Category:    msgCategory,
			Type:        msgType,
			Title:       msgTitle,
			Content:     msgContent,
			RelatedID:   attempt.ID,
			RelatedType: "attempt",
			IsRead:      false,
		}
		if err := tx.Create(msg).Error; err != nil {
			return resp, common.ErrNew(err, common.SysErr)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	resp = ResponseIS{
		ID:     attempt.ID,
		Status: attempt.Status,
	}

	return resp, nil
}

// ListByPhoto 获取某图片下的已审核答题记录
func (a *AttemptSvc) ListByPhoto(params PhotoAttemptsListParams) (resp AttemptForms, err error) {
	var attempts []model.Attempt
	var total int64
	// 查询已审核通过的答题记录，且排除未破解的记录
	query := model.DB.Model(&model.Attempt{}).
		Where("photo_id = ? AND status = ?", params.PhotoID, "solved")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	switch params.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "likes_count":
		query = query.Order("likes_count DESC")
	default:
		query = query.Order("created_at DESC")
	}
	if err := query.Preload("User").Preload("Photo.Activity").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]AttemptForm, 0, len(attempts))
	for _, at := range attempts {
		resp.List = append(resp.List, AttemptForm{
			ID: at.ID,
			Author: UserBrief{
				ID:        at.User.ID,
				Nickname:  at.User.Nickname,
				AvatarURL: at.User.AvatarURL,
			},
			Photo: PhotoBrief{
				ID:       at.Photo.ID,
				Title:    at.Photo.Title,
				ThumbURL: at.Photo.ThumbURL,
				Activity: ActivityBrief{ID: at.Photo.Activity.ID, Title: at.Photo.Activity.Title, Description: at.Photo.Activity.Description},
			},
			AnswerText: at.CommentText,
			ImageURL:   at.ImageURL,
			LikesCount: at.LikesCount,
			CreatedAt:  &at.CreatedAt,
		})
	}

	return resp, nil
}

// ListByPhotoUser 获取某图片下的用户答题记录
func (a *AttemptSvc) ListByPhotoUser(params PhotoAttemptsUserListParams) (resp UserAttemptForms, err error) {
	var attempts []model.Attempt
	var total int64
	// 查询答题记录
	query := model.DB.Model(&model.Attempt{}).
		Where("user_id = ? AND photo_id = ?", params.UserID, params.PhotoID)

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	switch params.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "likes_count":
		query = query.Order("likes_count DESC")
	default:
		query = query.Order("created_at DESC")
	}
	if err := query.Preload("User").Preload("Photo.Activity").
		Scopes(model.Paginate(params.PagerForm)).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]UserAttemptForm, 0, len(attempts))
	for _, at := range attempts {
		resp.List = append(resp.List, UserAttemptForm{
			ID: at.ID,
			Photo: PhotoBrief{
				ID:       at.Photo.ID,
				Title:    at.Photo.Title,
				ThumbURL: at.Photo.ThumbURL,
				Activity: ActivityBrief{ID: at.Photo.Activity.ID, Title: at.Photo.Activity.Title, Description: at.Photo.Activity.Description},
			},
			AnswerText:   at.CommentText,
			ImageURL:     at.ImageURL,
			Latitude:     at.Latitude,
			Longitude:    at.Longitude,
			LikesCount:   at.LikesCount,
			CreatedAt:    &at.CreatedAt,
			Status:       at.Status,
			RejectReason: at.RejectReason,
		})
	}

	return resp, nil
}

// ListUser 获取某用户的所有答题记录（个人主页用，支持按活动分段）
func (a *AttemptSvc) ListUser(params AttemptsListUserParams) (resp UserAttemptForms, err error) {

	var total int64
	var attempts []model.Attempt
	query := model.DB.Model(&model.Attempt{}).
		Where("user_id = ?", params.UserID)

	if params.ActivityID > 0 {
		query = query.Where("photo_id IN (SELECT id FROM photo WHERE activity_id = ?)", params.ActivityID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	switch params.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "likes_count":
		query = query.Order("likes_count DESC")
	default:
		query = query.Order("created_at DESC")
	}
	if err := query.Preload("Photo.Activity").Scopes(model.Paginate(params.PagerForm)).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp.Total = total
	resp.List = make([]UserAttemptForm, 0, len(attempts))
	for _, at := range attempts {
		resp.List = append(resp.List, UserAttemptForm{
			ID: at.ID,
			Photo: PhotoBrief{
				ID:       at.Photo.ID,
				Title:    at.Photo.Title,
				ThumbURL: at.Photo.ThumbURL,
				Activity: ActivityBrief{ID: at.Photo.Activity.ID, Title: at.Photo.Activity.Title, Description: at.Photo.Activity.Description},
			},
			AnswerText:   at.CommentText,
			ImageURL:     at.ImageURL,
			Latitude:     at.Latitude,
			Longitude:    at.Longitude,
			LikesCount:   at.LikesCount,
			CreatedAt:    &at.CreatedAt,
			Status:       at.Status,
			RejectReason: at.RejectReason,
		})
	}
	return resp, nil
}

// GetUserRank 获取用户在指定图片答题中的排名（按最早答对时间）
func (a *ActivitySvc) GetUserRank(userID int64, photoID int64) (rank int, err error) {
	// 1. 获取该用户最早答对时间
	var firstTime time.Time
	if err := model.DB.Model(&model.Attempt{}).
		Select("MIN(created_at)").
		Where("user_id = ? AND photo_id = ? AND status = ?", userID, photoID, "solved").
		Scan(&firstTime).Error; err != nil {
		return 0, common.ErrNew(err, common.SysErr)
	}
	if firstTime.IsZero() {
		return 0, nil // 未答对，无排名
	}

	// 2. 统计有多少不同用户的【最早答对时间】早于该用户
	if err := model.DB.Raw(
		"SELECT COUNT(DISTINCT user_id) + 1 FROM attempt WHERE status = ? AND photo_id = ? AND created_at < ?",
		"solved", photoID, firstTime).Scan(&rank).Error; err != nil {
		return 0, common.ErrNew(err, common.SysErr)
	}
	return rank, nil
}

// 计算两点距离
// earthRadius 地球平均半径，单位为米
const earthRadius = 6371000

// rad 将角度转换为弧度
func rad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

// DistanceBetweenGCJ02 计算两个GCJ-02坐标之间的距离，单位：米
func DistanceBetweenGCJ02(lat1, lng1, lat2, lng2 float64) float64 {
	// 1. 将纬度、经度从角度转为弧度
	radLat1 := rad(lat1)
	radLat2 := rad(lat2)

	// 2. 计算纬度和经度的差值（弧度）
	diffLat := radLat1 - radLat2
	diffLng := rad(lng1) - rad(lng2)

	// 3. 应用 Haversine 公式
	// a = sin²(Δlat/2) + cos(lat1) * cos(lat2) * sin²(Δlng/2)
	a := math.Pow(math.Sin(diffLat/2), 2) +
		math.Cos(radLat1)*math.Cos(radLat2)*
			math.Pow(math.Sin(diffLng/2), 2)

	// c = 2 * atan2(√a, √(1-a))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// 4. 距离 = 地球半径 * c
	distance := earthRadius * c
	return distance
}
