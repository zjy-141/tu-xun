package controller

import (
	"fmt"
	"net/http"
	"tu-xun/common"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Auth struct{}

// Register 用户注册
func (a *Auth) Register(c *gin.Context) {
	var params service.RegisterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		fmt.Printf("controller auth register: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	user, err := srv.Auth.Register(params)
	if err != nil {
		fmt.Printf("controller auth register: %v\n", err)
		c.Error(err)
		return
	}

	// 注册成功后自动登录
	SessionSet(c, "user-session", UserSession{
		ID:       user.ID,
		Username: user.Name,
		Level:    user.Level,
	})

	c.JSON(http.StatusCreated, ResponseNew(c, user))
}

// Login 用户登录
func (a *Auth) Login(c *gin.Context) {
	var params service.LoginParams
	if err := c.ShouldBindJSON(&params); err != nil {
		fmt.Printf("controller auth login: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	user, err := srv.Auth.Login(params)
	if err != nil {
		fmt.Printf("controller auth login: %v\n", err)
		c.Error(err)
		return
	}

	SessionSet(c, "user-session", UserSession{
		ID:       user.ID,
		Username: user.Name,
		Level:    user.Level,
	})

	c.JSON(http.StatusOK, ResponseNew(c, user))
}

// Logout 用户登出
func (a *Auth) Logout(c *gin.Context) {
	SessionClear(c)
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// Me 获取当前用户信息
func (a *Auth) Me(c *gin.Context) {
	session := SessionGet(c, "user-session")
	if session == nil {
		c.Error(common.ErrNew(fmt.Errorf("未登录"), common.AuthErr))
		return
	}
	us := session.(UserSession)

	user, err := srv.Auth.GetMe(us.ID)
	if err != nil {
		fmt.Printf("controller auth me: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, user))
}
