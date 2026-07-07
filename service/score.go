package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ScoreSvc struct{}

// MyScore 我的积分
func (s *ScoreSvc) MyScore(UserID int64) (resp ScoreTotal, err error) {

	var user model.User
	if err := model.DB.Model(&model.User{}).Where("id = ?", UserID).
		First(&user).Error; err != nil {
		return resp, err
	}
	resp.TotalScore = user.ScoreCount

	return resp, nil
}

// MyScoreLog 我的积分明细
func (s *ScoreSvc) MyScoreLog(params ScoreLogParams) (resp ScoreLogForms, err error) {

	var scoreLog []model.ScoreLog
	var total int64

	query := model.DB.Model(&model.ScoreLog{}).
		Where("User_id = ?", params.UserID)

	if err := query.Count(&total).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := query.Scopes(model.Paginate(params.PagerForm)).
		Find(&scoreLog).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp.Total = total
	resp.ScoreLogs = make([]ScoreLogForm, 0, len(scoreLog))
	for _, sl := range scoreLog {
		resp.ScoreLogs = append(resp.ScoreLogs, ScoreLogForm{
			ID:          sl.ID,
			Delta:       sl.Delta,
			Balance:     sl.Balance,
			Reason:      sl.Reason,
			RelatedID:   sl.RelatedID,
			RelatedType: sl.RelatedType,
			CreatedAt:   sl.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}

// // TxRegularScoreReward 事务常用积分变化
// func (s *ScoreSvc) TxRegularScoreChange(params ScoreChangeParams) (resp ResponseIS, err error) {
// 	tx := model.DB.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 			panic(r)
// 		}
// 	}()

// 	// 查询用户
// 	var user model.User
// 	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
// 		Select("id", "score_count").
// 		Where("id = ?", params.UserID).
// 		First(&user).Error; err != nil {
// 		tx.Rollback()
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return resp, common.ErrNew(errors.New("用户不存在"), common.ParamErr)
// 		}
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	// 计算新余额
// 	newBalance := user.ScoreCount + params.Delta
// 	if newBalance < 0 {
// 		tx.Rollback()
// 		return resp, common.ErrNew(errors.New("积分余额不足"), common.ParamErr)
// 	}

// 	// 更新用户积分
// 	if err := tx.Model(&user).Update("score_count", newBalance).Error; err != nil {
// 		tx.Rollback()
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	// 新建积分日志
// 	scoreLog := &model.ScoreLog{
// 		UserID:      params.UserID,
// 		Delta:       params.Delta,
// 		Balance:     newBalance,
// 		Reason:      params.Reason,
// 		RelatedID:   params.RelatedID,
// 		RelatedType: params.RelatedType,
// 		Remark:      params.Remark,
// 	}

// 	if err := tx.Create(scoreLog).Error; err != nil {
// 		tx.Rollback()
// 		return resp, common.ErrNew(err, common.SysErr)
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		return resp, common.ErrNew(errors.New("事务提交失败"), common.SysErr)
// 	}

// 	resp = ResponseIS{
// 		ID:     scoreLog.ID,
// 		Status: "success",
// 	}
// 	return resp, nil
// }

// RegularScoreReward 常用积分变化
func (s *ScoreSvc) RegularScoreChange(tx *gorm.DB, params ScoreChangeParams) (resp ResponseIS, err error) {
	// 查询用户
	var user model.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "score_count").
		Where("id = ?", params.UserID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, errors.New("用户不存在")
		}
		return resp, err
	}

	// 计算新余额
	newBalance := user.ScoreCount + params.Delta
	if newBalance < 0 {
		return resp, errors.New("积分余额不足")
	}

	// 更新用户积分
	if err := tx.Model(&user).Update("score_count", newBalance).Error; err != nil {
		return resp, err
	}
	// 新建积分日志
	scoreLog := &model.ScoreLog{
		UserID:      params.UserID,
		Delta:       params.Delta,
		Balance:     newBalance,
		Reason:      params.Reason,
		RelatedID:   params.RelatedID,
		RelatedType: params.RelatedType,
		Remark:      params.Remark,
	}
	if err := tx.Create(scoreLog).Error; err != nil {
		return resp, err
	}
	resp = ResponseIS{
		ID:     scoreLog.ID,
		Status: "success",
	}
	return resp, nil
}
