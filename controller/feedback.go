package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Feedback struct {
}

// Create 发送反馈（multipart/form-data）
func (f *Feedback) Create(c *gin.Context) {
	var params service.FeedbackCreateParams

	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller feedback create: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	resp, err := srv.FeedbackSvc.Create(params)
	if err != nil {
		logger.Errorf("service feedback create: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// List 获取当前用户的反馈列表
func (f *Feedback) List(c *gin.Context) {
	var params service.FeedbackListParams

	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Errorf("controller feedback list: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.FeedbackSvc.List(params)
	if err != nil {
		logger.Errorf("service feedback list: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Detail 获取反馈详情
func (f *Feedback) Detail(c *gin.Context) {
	var params service.FeedbackGetByIDParams
	feedbackID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || feedbackID <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	params.FeedbackID = feedbackID

	resp, err := srv.FeedbackSvc.Detail(params)
	if err != nil {
		logger.Errorf("service feedback detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Review 获取反馈详情
func (f *Feedback) Review(c *gin.Context) {
	var params service.FeedbackReviewParams
	feedbackID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || feedbackID <= 0 {
		logger.Errorf("controller feedback review: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller feedback review: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.FeedbackID = feedbackID

	resp, err := srv.FeedbackSvc.Review(params)
	if err != nil {
		logger.Errorf("service feedback detail: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
