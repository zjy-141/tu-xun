package controller

import (
	"fmt"
	"net/http"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Admin struct{}

// PendingPhotos 获取待审核图片列表
func (a *Admin) PendingPhotos(c *gin.Context) {
	var params service.PendingPhotosParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.Limit = 10
	}

	data, err := srv.Admin.PendingPhotos(params)
	if err != nil {
		fmt.Printf("controller admin pending photos: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}

// ReviewPhoto 审核图片
func (a *Admin) ReviewPhoto(c *gin.Context) {
	var params service.ReviewPhotoParams
	if err := c.ShouldBindUri(&params); err != nil {
		fmt.Printf("controller admin review photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		fmt.Printf("controller admin review photo: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	data, err := srv.Admin.ReviewPhoto(params)
	if err != nil {
		fmt.Printf("controller admin review photo: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}

// PendingAttempts 获取待审核答题记录
func (a *Admin) PendingAttempts(c *gin.Context) {
	var params service.PendingAttemptsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.Limit = 10
	}

	data, err := srv.Admin.PendingAttempts(params)
	if err != nil {
		fmt.Printf("controller admin pending attempts: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}

// ReviewAttempt 审核答题记录
func (a *Admin) ReviewAttempt(c *gin.Context) {
	var params service.ReviewAttemptParams
	if err := c.ShouldBindUri(&params); err != nil {
		fmt.Printf("controller admin review attempt: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		fmt.Printf("controller admin review attempt: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	data, err := srv.Admin.ReviewAttempt(params)
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
		ID int64 `uri:"id" binding:"min=1"`
	}
	if err := c.ShouldBindUri(&uriForm); err != nil {
		fmt.Printf("controller admin claim prize: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	data, err := srv.Admin.ClaimPrize(uriForm.ID)
	if err != nil {
		fmt.Printf("controller admin claim prize: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}
