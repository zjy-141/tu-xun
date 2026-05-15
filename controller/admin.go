package controller

import (
	"fmt"
	"net/http"
	"tu-xun/common"

	"github.com/gin-gonic/gin"
)

type Admin struct{}

// PendingPhotos 获取待审核图片列表
func (a *Admin) PendingPhotos(c *gin.Context) {
	var form common.PagerForm
	if err := c.ShouldBindQuery(&form); err != nil {
		form.Page = 1
		form.Limit = 10
	}

	data, err := srv.Admin.PendingPhotos(form.Page, form.Limit)
	if err != nil {
		fmt.Printf("controller admin pending photos: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}

// ReviewPhoto 审核图片
func (a *Admin) ReviewPhoto(c *gin.Context) {
	var uriForm common.IDUriForm
	if err := c.ShouldBindUri(&uriForm); err != nil {
		fmt.Printf("controller admin review photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	var body struct {
		Action       string `json:"action" binding:"required"`
		RejectReason string `json:"reject_reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Printf("controller admin review photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	data, err := srv.Admin.ReviewPhoto(int64(uriForm.ID), body.Action, body.RejectReason)
	if err != nil {
		fmt.Printf("controller admin review photo: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}

// PendingAttempts 获取待审核答题记录
func (a *Admin) PendingAttempts(c *gin.Context) {
	var form common.PagerForm
	if err := c.ShouldBindQuery(&form); err != nil {
		form.Page = 1
		form.Limit = 10
	}

	data, err := srv.Admin.PendingAttempts(form.Page, form.Limit)
	if err != nil {
		fmt.Printf("controller admin pending attempts: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}

// ReviewAttempt 审核答题记录
func (a *Admin) ReviewAttempt(c *gin.Context) {
	var uriForm struct {
		ID int `uri:"id" binding:"min=1"`
	}
	if err := c.ShouldBindUri(&uriForm); err != nil {
		fmt.Printf("controller admin review attempt: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	var body struct {
		Action       string `json:"action" binding:"required"`
		RejectReason string `json:"reject_reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Printf("controller admin review attempt: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	data, err := srv.Admin.ReviewAttempt(int64(uriForm.ID), body.Action, body.RejectReason)
	if err != nil {
		fmt.Printf("controller admin review attempt: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}

// ClaimPrize 标记奖品已发放
func (a *Admin) ClaimPrize(c *gin.Context) {
	var uriForm struct {
		ID int `uri:"id" binding:"min=1"`
	}
	if err := c.ShouldBindUri(&uriForm); err != nil {
		fmt.Printf("controller admin claim prize: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	data, err := srv.Admin.ClaimPrize(int64(uriForm.ID))
	if err != nil {
		fmt.Printf("controller admin claim prize: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}
