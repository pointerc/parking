package service

import (
	"github.com/gin-gonic/gin"
	"parking/goapp/handle/login_handle"
)

type ILogin interface {
	LoginSystem(c *gin.Context)
	Register(c *gin.Context)
	LoginRouter(router *gin.Engine)
}

type Login struct{
	login login_handle.Login
}

func (l *Login) LoginRouter(router *gin.Engine) {
	//router.GET("/system/login/1", l.LoginSystem)
	//router.GET("/system/register/1", l.Register)
	handle := router.Group("/system")
	handle.POST("/login/1", l.LoginSystem)
	handle.POST("/register/1", l.Register)
}

func (l *Login) LoginSystem(c *gin.Context) {
	l.login.UserLogin(c)
}

func (l *Login) Register(c *gin.Context) {
	l.login.Register(c)
}
