package controller

import (
	"fmt"
	"net/http"
	"tu-xun/common"

	"github.com/gin-gonic/gin"
)

type Attempt struct{}

// Submit 提交答题
func (a *Attempt) Submit(c *gin.Context) {
	us := getCurrentUser(c)
	if us == nil {
		return
	}

	var uriForm common.IDUriForm
	if err := c.ShouldBindUri(&uriForm); err != nil {
		fmt.Printf("controller attempt submit: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	guessedLocation := c.PostForm("guessed_location")
	if guessedLocation == "" {
		c.Error(common.ErrNew(fmt.Errorf("猜测地点不能为空"), common.ParamErr))
		return
	}

	imageFile, err := c.FormFile("image")
	if err != nil {
		c.Error(common.ErrNew(fmt.Errorf("请上传匹配照片"), common.ParamErr))
		return
	}

	attempt, err := srv.Attempt.Create(int64(uriForm.ID), int64(us.ID), guessedLocation, imageFile)
	if err != nil {
		fmt.Printf("controller attempt submit: %v\n", err)
		c.Error(err)
		return
	}

	msg := "已提交，等待管理员审核。若审核通过且本题尚未被破解，您将获得奖品。"
	c.JSON(http.StatusCreated, ResponseNew(c, map[string]any{
		"attempt_id": attempt.ID,
		"photo_id":   attempt.PhotoID,
		"status":     attempt.Status,
		"message":    msg,
	}))
}

// MyAttempts 获取我对某图片的所有答题记录
func (a *Attempt) MyAttempts(c *gin.Context) {
	us := getCurrentUser(c)
	if us == nil {
		return
	}

	var uriForm common.IDUriForm
	if err := c.ShouldBindUri(&uriForm); err != nil {
		fmt.Printf("controller attempt my: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	data, err := srv.Attempt.MyAttempts(int64(uriForm.ID), int64(us.ID))
	if err != nil {
		fmt.Printf("controller attempt my: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, data))
}
