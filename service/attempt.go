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

const maxAttemptsPerPhoto = 5

// Create 提交作答
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
			return resp, common.ErrNew(errors.New("图片不存在"), common.OpErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}
	if photo.Status != "approved" {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该图片尚未通过审核，暂不可答题"), common.OpErr)
	}

	// 题目作者不能作答自己投稿的题目
	if photo.UserID == info.UserID {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("不能作答自己投稿的题目"), common.OpErr)
	}

	// 活动已结束不可作答
	activitySvc := ActivitySvc{}
	if active, _ := activitySvc.IsActivityActive(photo.ActivityID); !active {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该活动不在进行中"), common.ParamErr)
	}

	// 已破解成功禁止再次作答
	var solvedCount int64
	tx.Model(&model.Attempt{}).Where("photo_id = ? AND user_id = ? AND status = ?", info.PhotoID, info.UserID, "solved").Count(&solvedCount)
	if solvedCount > 0 {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("已破解成功，不可再次作答"), common.OpErr)
	}

	// 作答次数上限检查
	var userAttemptsCount int64
	tx.Model(&model.Attempt{}).Where("photo_id = ? AND user_id = ?", info.PhotoID, info.UserID).Count(&userAttemptsCount)
	if userAttemptsCount >= maxAttemptsPerPhoto {
		tx.Rollback()
		return resp, common.ErrNew(fmt.Errorf("作答次数已达上限（%d次）", maxAttemptsPerPhoto), common.OpErr)
	}

	// 保存答题图片（仅缩略图）
	imageURL, imgW, imgH, err := saveThumbnailOnly(info.ImageFile, "attempts")
	if err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	gcjLat, gcjLng := WGS84orGCJ02ToGCJ02(info.Latitude, info.Longitude, info.CoordType)

	status := "pending"
	rejectReason := ""
	var reviewedAt *time.Time
	if config.Config.AUTO_APPROVAL == "attemptAndComment" || config.Config.AUTO_APPROVAL == "all" {
		distance := DistanceBetweenGCJ02(photo.Latitude, photo.Longitude, gcjLat, gcjLng)
		now := time.Now()
		if distance <= 50 {
			status = "solved"
		} else {
			status = "unsolved"
			rejectReason = "作答位置不在题目范围内，无法通过审核"
		}
		reviewedAt = &now
	}

	attempt := &model.Attempt{
		PhotoID:      info.PhotoID,
		UserID:       info.UserID,
		ImageURL:     imageURL,
		ImageWidth:   imgW,
		ImageHeight:  imgH,
		Latitude:     gcjLat,
		Longitude:    gcjLng,
		CoordType:    info.CoordType,
		LikesCount:   0,
		Status:       status,
		RejectReason: rejectReason,
		ReviewedAt:   reviewedAt,
	}

	if err := tx.Create(attempt).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 更新作答次数
	if err := tx.Model(&model.Photo{}).
		Where("id = ?", attempt.PhotoID).
		Update("attempts_count", gorm.Expr("attempts_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 自动审核：发放积分和更新破解数
	if status == "solved" {
		if err := tx.Model(&model.Photo{}).Where("id = ?", attempt.PhotoID).
			Update("solved_count", gorm.Expr("solved_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			return resp, common.ErrNew(err, common.SysErr)
		}

		if attempt.UserID != photo.UserID {
			// 计算排名（事务内可见当前未提交的 solved 记录）
			rank, err := activitySvc.GetUserRank(tx, attempt.UserID, attempt.PhotoID)
			if err != nil {
				tx.Rollback()
				return resp, common.ErrNew(err, common.SysErr)
			}

			// 优先使用活动奖励阶梯，未配置则使用默认阶梯
			delta := calcScoreByRank(rank)
			tiers := photo.Activity.AttemptRewardTiers
			if len(tiers) > 0 {
				sort.Slice(tiers, func(i, j int) bool {
					return tiers[i].Batch < tiers[j].Batch
				})
				for _, tier := range tiers {
					if rank <= tier.RankLimit {
						delta = tier.AttemptPoints
						break
					}
				}
			}

			scoreSvc := ScoreSvc{}
			if _, err := scoreSvc.RegularScoreChange(tx, ScoreChangeParams{
				UserID:      attempt.UserID,
				Delta:       delta,
				Reason:      "answer_correct",
				RelatedID:   attempt.PhotoID,
				RelatedType: "photo",
				Remark:      fmt.Sprintf("答对题目，排名第 %d，获得 %d 积分", rank, delta),
			}); err != nil {
				tx.Rollback()
				return resp, common.ErrNew(err, common.SysErr)
			}
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

// calcScoreByRank 按排名计算积分（简化默认阶梯）
func calcScoreByRank(rank int) int {
	switch {
	case rank <= 3:
		return 20
	case rank <= 10:
		return 15
	default:
		return 10
	}
}

// ListSolves 获取某图片下的破解记录（仅 solved，公开）
func (a *AttemptSvc) ListSolves(params SolvesListParams, userID int64) (resp SolveItemPage, err error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{}).
		Where("photo_id = ? AND status = ?", params.PhotoID, "solved").
		Order("created_at DESC")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("User").
		Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 批量查询点赞状态
	attemptIDs := make([]int64, len(attempts))
	for i, at := range attempts {
		attemptIDs[i] = at.ID
	}
	likedSet := make(map[int64]bool)
	if userID > 0 && len(attemptIDs) > 0 {
		var likes []model.Like
		model.DB.Where("user_id = ? AND target_type = ? AND target_id IN ?", userID, "attempt", attemptIDs).Find(&likes)
		for _, lk := range likes {
			likedSet[lk.TargetID] = true
		}
	}

	resp.Total = total
	resp.List = make([]SolveItem, 0, len(attempts))
	for _, at := range attempts {
		resp.List = append(resp.List, SolveItem{
			ID: at.ID,
			Author: UserBrief{
				ID:       at.User.ID,
				Nickname: at.User.Nickname,
				Avatar:   at.User.AvatarURL,
			},
			Image: Media{
				ThumbURL: at.ImageURL,
				Width:    at.ImageWidth,
				Height:   at.ImageHeight,
			},
			Location: Location{
				Longitude: at.Longitude,
				Latitude:  at.Latitude,
				CoordType: at.CoordType,
			},
			LikesCount: at.LikesCount,
			Liked:      likedSet[at.ID],
			CreatedAt:  &at.CreatedAt,
		})
	}

	return resp, nil
}

// ListByPhotoUser 获取某图片下当前用户的作答记录
func (a *AttemptSvc) ListByPhotoUser(userID int64, photoID int64) (resp AttemptRecordPage, err error) {
	var attempts []model.Attempt
	var total int64

	query := model.DB.Model(&model.Attempt{}).
		Where("user_id = ? AND photo_id = ?", userID, photoID).
		Order("created_at DESC")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]AttemptRecord, 0, len(attempts))
	for _, at := range attempts {
		resp.List = append(resp.List, AttemptRecord{
			ID: at.ID,
			Image: Media{
				ThumbURL: at.ImageURL,
				Width:    at.ImageWidth,
				Height:   at.ImageHeight,
			},
			Location: Location{
				Longitude: at.Longitude,
				Latitude:  at.Latitude,
				CoordType: at.CoordType,
			},
			Status:       at.Status,
			RejectReason: at.RejectReason,
			CreatedAt:    &at.CreatedAt,
		})
	}

	return resp, nil
}

// ListUser 获取用户的所有作答记录
func (a *AttemptSvc) ListUser(params AttemptsListUserParams) (resp UserAttemptCardPage, err error) {
	var total int64
	var attempts []model.Attempt

	query := model.DB.Model(&model.Attempt{}).
		Where("user_id = ?", params.UserID).
		Order("created_at DESC, id DESC")

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Preload("Photo").
		Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).
		Find(&attempts).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.List = make([]UserAttemptCard, 0, len(attempts))
	for _, at := range attempts {
		// 计算该题的用户作答次数
		var uac int64
		model.DB.Model(&model.Attempt{}).Where("photo_id = ? AND user_id = ?", at.PhotoID, params.UserID).Count(&uac)

		resp.List = append(resp.List, UserAttemptCard{
			ID: at.ID,
			Photo: PhotoBrief{
				ID:    at.Photo.ID,
				Title: at.Photo.Title,
				Image: Media{ThumbURL: at.Photo.ThumbURL, Width: at.Photo.ThumbWidth, Height: at.Photo.ThumbHeight},
			},
			UserAttemptsCount: int(uac),
			Status:            at.Status,
			CreatedAt:         &at.CreatedAt,
		})
	}

	return resp, nil
}

// GetUserRank 获取用户在指定图片答题中的排名（按最早答对时间）
func (a *ActivitySvc) GetUserRank(db *gorm.DB, userID int64, photoID int64) (rank int, err error) {
	var firstTime time.Time
	if err := db.Model(&model.Attempt{}).
		Select("MIN(created_at)").
		Where("user_id = ? AND photo_id = ? AND status = ?", userID, photoID, "solved").
		Scan(&firstTime).Error; err != nil {
		return 0, common.ErrNew(err, common.SysErr)
	}
	if firstTime.IsZero() {
		return 0, nil
	}

	if err := db.Raw(
		"SELECT COUNT(DISTINCT user_id) + 1 FROM attempt WHERE status = ? AND photo_id = ? AND created_at < ?",
		"solved", photoID, firstTime).Scan(&rank).Error; err != nil {
		return 0, common.ErrNew(err, common.SysErr)
	}
	return rank, nil
}

// earthRadius 地球平均半径，单位为米
const earthRadius = 6371000

func rad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

// DistanceBetweenGCJ02 计算两个GCJ-02坐标之间的距离，单位：米
func DistanceBetweenGCJ02(lat1, lng1, lat2, lng2 float64) float64 {
	radLat1 := rad(lat1)
	radLat2 := rad(lat2)
	diffLat := radLat1 - radLat2
	diffLng := rad(lng1) - rad(lng2)
	a := math.Pow(math.Sin(diffLat/2), 2) +
		math.Cos(radLat1)*math.Cos(radLat2)*
			math.Pow(math.Sin(diffLng/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}
