package controller

import (
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Announcement struct{}

// List 客户端通知列表（含未读数）
func (ctr *Announcement) List(c *gin.Context) {
	var params service.AnnouncementListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.AnnouncementSvc.List(userID, params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// Detail 客户端通知详情（读取即标记已读）
func (ctr *Announcement) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userID := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.AnnouncementSvc.GetByID(userID, id)
	if err != nil {
		c.Error(err)
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, ResponseNew(c, nil))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// AdminList 管理端通知列表（含已读人数）
func (ctr *Announcement) AdminList(c *gin.Context) {
	var params service.AdminAnnouncementListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AnnouncementSvc.AdminList(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// AdminDetail 管理端通知详情（不标记已读）
func (ctr *Announcement) AdminDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AnnouncementSvc.AdminGetByID(id)
	if err != nil {
		c.Error(err)
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, ResponseNew(c, nil))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// AdminCreate 管理员发布通知（multipart/form-data）
func (ctr *Announcement) AdminCreate(c *gin.Context) {
	var params service.CreateAnnouncementRequest
	if err := c.ShouldBind(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AnnouncementSvc.Create(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// AdminUpdate 管理员更新通知（multipart/form-data）
func (ctr *Announcement) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var params service.UpdateAnnouncementRequest
	if err := c.ShouldBind(&params); err != nil {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	resp, err := srv.AnnouncementSvc.Update(id, params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// AdminDelete 管理员删除通知
func (ctr *Announcement) AdminDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	if err := srv.AnnouncementSvc.Delete(id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}
