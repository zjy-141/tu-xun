package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Score struct{}

// MyScoreLog 我的积分明细
func (ctr *Score) MyScoreLog(c *gin.Context) {
	var params service.ScoreLogParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.ScoreSvc.MyScoreLog(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
