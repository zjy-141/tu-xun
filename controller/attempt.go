package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Attempt struct{}

// Submit 提交答题
func (a *Attempt) Submit(c *gin.Context) {

	var params service.AttemptCreateParams
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || photoID <= 0 {
		logger.Errorf("controller attempt submit: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller attempt submit: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.PhotoID = photoID
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	resp, err := srv.AttemptSvc.Create(params)
	if err != nil {
		logger.Errorf("service attempt submit: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// // UserAttempt 获取某用户的所有答题记录（个人主页用）
// func (a *Attempt) UserAttempt(c *gin.Context) {
// 	var params service.AttemptShowParams
// 	if err := c.ShouldBindQuery(&params); err != nil {
// 		logger.Errorf("controller attempt show: %v\n", err)
// 		c.Error(common.ErrNew(err, common.ParamErr))
// 		return
// 	}
// 	NetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil || NetID <= 0 {
// 		logger.Errorf("controller attempt show: %v\n", err)
// 		c.Error(common.ErrNew(err, common.ParamErr))
// 		return
// 	}
// 	params.NetID = NetID

// 	resp, err := srv.Attempt.AttemptShow(params)
// 	if err != nil {
// 		logger.Errorf("controller attempt show	: %v\n", err)
// 		c.Error(err)
// 		return
// 	}

// 	c.JSON(http.StatusOK, ResponseNew(c, resp))
// }
