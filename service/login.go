package service

import "github.com/gin-gonic/gin"

type ILogin interface {
	LoginRouter(router *gin.Engine)
}

type Login struct{

}

func (l *Login) LoginRouter(router *gin.Engine) {
	handle := router.Group("/system")
	handle.GET("/login/1", )
}

func (l *Login) LoginSystem(c *gin.Context) {

}
