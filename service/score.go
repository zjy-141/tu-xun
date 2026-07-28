package service

import (
	"errors"

	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ScoreSvc struct{}

// MyScoreLog 我的积分明细
func (s *ScoreSvc) MyScoreLog(params ScoreLogParams) (ScoreLogPage, error) {
	var scoreLogs []model.ScoreLog
	var total int64

	query := model.DB.Model(&model.ScoreLog{}).
		Where("user_id = ?", params.UserID).
		Order("id DESC")

	if err := query.Count(&total).Error; err != nil {
		return ScoreLogPage{}, common.ErrNew(err, common.SysErr)
	}

	if err := query.Scopes(model.Paginate(params.PagerForm)).
		Find(&scoreLogs).Error; err != nil {
		return ScoreLogPage{}, common.ErrNew(err, common.SysErr)
	}

	resp := ScoreLogPage{
		Total: total,
		List:  make([]ScoreLogItem, 0, len(scoreLogs)),
	}
	for _, sl := range scoreLogs {
		resp.List = append(resp.List, ScoreLogItem{
			ID:          sl.ID,
			Delta:       sl.Delta,
			Balance:     sl.Balance,
			Reason:      sl.Reason,
			RelatedID:   sl.RelatedID,
			RelatedType: sl.RelatedType,
			CreatedAt:   &sl.CreatedAt,
		})
	}
	return resp, nil
}

// RegularScoreChange 常用积分变化（在事务中执行，由调用方传入 tx 和参数）
func (s *ScoreSvc) RegularScoreChange(tx *gorm.DB, params ScoreChangeParams) (ResponseIS, error) {
	// 查询用户
	var user model.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "score_count").
		Where("id = ?", params.UserID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ResponseIS{}, errors.New("用户不存在")
		}
		return ResponseIS{}, err
	}

	// 计算新余额
	newBalance := user.ScoreCount + params.Delta
	if newBalance < 0 {
		return ResponseIS{}, errors.New("积分余额不足")
	}

	// 更新用户积分
	if err := tx.Model(&user).Update("score_count", newBalance).Error; err != nil {
		return ResponseIS{}, err
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
		return ResponseIS{}, err
	}

	return ResponseIS{
		ID:     scoreLog.ID,
		Status: "success",
	}, nil
}
