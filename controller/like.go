package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"

	"github.com/gin-gonic/gin"
)

type Like struct{}

// SetPhotoLike 幂等切换图片点赞状态
func (ctr *Like) SetPhotoLike(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.LikeSvc.SetLike(userID, "photo", id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// SetCommentLike 幂等切换评论点赞状态
func (ctr *Like) SetCommentLike(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.LikeSvc.SetLike(userID, "comment", id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// SetSolveLike 幂等切换破解记录点赞状态
func (ctr *Like) SetSolveLike(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.LikeSvc.SetLike(userID, "attempt", id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
