package service

import (
	"tu-xun/common"
	"tu-xun/model"
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
