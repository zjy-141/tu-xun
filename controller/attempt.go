package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Attempt struct{}

// Submit 提交答题
func (a *Attempt) Submit(c *gin.Context) {

	var params service.CreateAttemptParams
	if err := c.ShouldBindUri(&params); err != nil {
		logger.Errorf("controller attempt submit: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller attempt submit: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	attempt, err := srv.Attempt.Create(params)
	if err != nil {
		logger.Errorf("controller attempt submit: %v\n", err)
		c.Error(err)
		return
	}

	msg := "已提交，等待管理员审核。若审核通过且本题尚未被破解，您将获得奖品。"
	c.JSON(http.StatusCreated, ResponseNew(c, service.SubmitAttemptResponse{
		AttemptID: attempt.ID,
		PhotoID:   attempt.PhotoID,
		Status:    attempt.Status,
		Message:   msg,
	}))
}

// MyAttempts 获取我对某图片的所有答题记录
func (a *Attempt) MyAttempts(c *gin.Context) {

	var params service.MyAttemptsParams
	if err := c.ShouldBindUri(&params); err != nil {
		logger.Errorf("controller attempt my: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	resp, err := srv.Attempt.MyAttempts(params)
	if err != nil {
		logger.Errorf("controller attempt my: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
