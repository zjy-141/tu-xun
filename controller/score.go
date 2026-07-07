package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Score struct{}

// MyScore 我的积分
func (s *Score) MyScore(c *gin.Context) {

	UserID := SessionGet(c, "user-session").(UserSession).ID

	resp, err := srv.ScoreSvc.MyScore(UserID)
	if err != nil {
		logger.Errorf("service score my score: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// MyScoreLog 我的积分明细
func (s *Score) MyScoreLog(c *gin.Context) {
	var params service.ScoreLogParams

	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller score my score log: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	resp, err := srv.ScoreSvc.MyScoreLog(params)
	if err != nil {
		logger.Errorf("service score my score log: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
