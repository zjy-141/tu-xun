package controller

import (
	"net/http"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Exchange struct{}

// Claim 兑换奖品
func (e *Exchange) Claim(c *gin.Context) {

	var params service.ExchangeCreateParams
	// goodID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	// if err != nil || goodID <= 0 {
	// 	logger.Errorf("controller comment create: %v\n", err)
	// 	c.Error(common.ErrNew(err, common.ParamErr))
	// 	return
	// }
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller exchange claim: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	// params.GoodID = goodID
	params.UserID = SessionGet(c, "user-session").(UserSession).ID
		params.IdempotencyKey = c.GetHeader("Idempotency-Key")

	resp, err := srv.ExchangeSvc.Claim(params)
	if err != nil {
		logger.Errorf("service exchange claim: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// List 获取兑奖记录
func (e *Exchange) List(c *gin.Context) {
	var params service.ExchangeListParams

	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller exchange list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.ExchangeSvc.List(params)
	if err != nil {
		logger.Errorf("service exchange list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
