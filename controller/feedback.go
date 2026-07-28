package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Feedback struct {
}

// Create 发送反馈（multipart/form-data）
func (ctr *Feedback) Create(c *gin.Context) {
	var params service.FeedbackCreateParams
	if err := c.ShouldBind(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.FeedbackSvc.Create(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, ResponseNew(c, resp))
}

// List 获取反馈列表
func (ctr *Feedback) List(c *gin.Context) {
	var params service.FeedbackListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.FeedbackSvc.List(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Detail 获取反馈详情
func (ctr *Feedback) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.FeedbackSvc.Detail(id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Review 处理反馈（更新状态）
func (ctr *Feedback) Review(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var params service.FeedbackReviewParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.FeedbackID = id
	if err := srv.FeedbackSvc.Review(params); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}
