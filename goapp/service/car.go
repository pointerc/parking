package service

import (
	"github.com/gin-gonic/gin"
	"parking/goapp/handle/car_handle"
	"parking/goapp/middlewaire"
)

type ICar interface {
	VehicleAdmission(c *gin.Context)
	VehicleAppearance(c *gin.Context)
	CarStop(router *gin.Engine)
}

type Car struct {
	car car_handle.Car
}

func (c *Car) CarStop(router *gin.Engine) {
	handle := router.Group("/system/car")
	middle := middlewaire.Token{}
	handle.Use(middle.CheckToken())
	{
		handle.POST("/vehicle/admission/1", c.VehicleAdmission)		//车辆入场
		handle.POST("/vehicle/appearance/1", c.VehicleAppearance)	//车辆出场
	}
}

func (car *Car) VehicleAdmission(c *gin.Context) {
	car.car.VehileAdmission(c)
}

func (car *Car) VehicleAppearance(c *gin.Context) {
	car.car.VehileAppearance(c)
}
