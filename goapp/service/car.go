package service

import (
	"github.com/gin-gonic/gin"
	"parking/goapp/handle/car_handle"
	"parking/goapp/middlewaire"
)

type ICar interface {
	VehicleAdmission(c *gin.Context)
	VehicleAppearance(c *gin.Context)
	GetParkInfo(c *gin.Context)
	GetParkPrice(c *gin.Context)
	GetHistoryInfo(c *gin.Context)
	GetNowParkInfo(c *gin.Context)
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
		handle.POST("/get/parking/info/1", c.GetParkInfo)				//当前停车场信息
		handle.POST("/get/parking/price/1", c.GetParkPrice)			//查询
		handle.POST("/get/history/1", c.GetHistoryInfo)				//查询用户历史停车记录
		handle.POST("/get/now/parkinfo/1", c.GetNowParkInfo)
	}
}

func (car *Car) VehicleAdmission(c *gin.Context) {
	car.car.VehileAdmission(c)
}

func (car *Car) VehicleAppearance(c *gin.Context) {
	car.car.VehileAppearance(c)
}

func (car *Car) GetParkInfo(c *gin.Context) {
	car.car.GetParkInfo(c)
}

func (car *Car) GetParkPrice(c *gin.Context) {
	car.car.GetParkPrice(c)
}

func (car *Car)GetHistoryInfo(c *gin.Context)  {
	car.car.GetUserHistory(c)
}

func (car *Car)GetNowParkInfo(c *gin.Context) {
	car.car.GetNowParkInfo(c)
}
