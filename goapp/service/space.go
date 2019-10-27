package service

import (
	"github.com/gin-gonic/gin"
	"parking/goapp/handle/space_handle"
	"parking/goapp/middlewaire"
)

type ISpace interface {
	SpaceInit(c *gin.Context)
	FreeSpace(c *gin.Context)
	SpaceRoter(router *gin.Engine)
}

type Space struct {
	space space_handle.Space
}

func (s *Space) SpaceRoter(router *gin.Engine) {
	handle := router.Group("/system/park")
	middle := middlewaire.Token{}
	handle.Use(middle.CheckToken())
	{
		handle.GET("/init/1", s.SpaceInit)
		handle.GET("/space/free/1", s.FreeSpace)
	}
}

func (s *Space) SpaceInit(c *gin.Context) {
	s.space.SpaceInit(c)
}

func (s *Space) FreeSpace(c *gin.Context) {
	s.space.FreeSpace(c)
}
