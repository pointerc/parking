package car_handle

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"parking/comm/db"
	"parking/comm/response"
	"parking/goapp/model/car_model"
	"time"
)

type Car struct {}

//车辆入场
func (car *Car) VehileAdmission(c *gin.Context) {
	var (
		uid string
		carInfo car_model.CarInfo
		match bson.M
		ParkInfo struct {
			ID bson.ObjectId `bson:"_id"`
			Status int `bson:"status"`
				 }
		resp = &response.JsonData{}
		err error
	)

	uid = c.GetString("uid")
	if !bson.IsObjectIdHex(uid) {
		resp.ExecFail(c, "用户ID异常")
		return
	}

	err = c.ShouldBindJSON(&carInfo)
	if err != nil {
		resp.ExecFail(c, "解析前端请求失败")
		return
	}

	//根据车位号，车类型查询车位id
	match = bson.M{"place_no": carInfo.PlaceNo, "car_type": carInfo.CarType}
	err = db.Mgo.Collection("Park", func(c *mgo.Collection) error {
		return c.Find(match).One(&ParkInfo)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "系统异常")
		return
	}
	if ParkInfo.ID.Hex() == "" {
		resp.ExecFail(c, "不存在该类型的车位")
		return
	}

	if ParkInfo.Status == 2 {
		resp.ExecFail(c, "该车位已被占用，请使用其他同类型的车位")
		return
	}

	//车辆进入，创建车辆信息
	carInfo.ID = bson.NewObjectId()
	carInfo.SpaceId = ParkInfo.ID
	carInfo.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	carInfo.StartTime = time.Now().Unix()
	carInfo.EndTime = 0
	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Insert(carInfo)
	})

	//更新车位状态
	if err != nil {
		resp.ExecFail(c, "车辆入场失败")
		return
	}

	err = db.Mgo.Collection("Park", func(c *mgo.Collection) error {
		return c.Update(bson.M{"_id": ParkInfo.ID}, bson.M{"$set": bson.M{"status": 1}})
	})
	if err != nil {
		resp.ExecFail(c, "更新车位状态失败")
		return
	}
	resp.Succ(c)
}

//车辆出场
func (car *Car) VehileAppearance(c *gin.Context) {
	var (
		uid string
		match bson.M
		carInfo car_model.CarInfo
		stopInfo car_model.CarInfo
		carStopInfo car_model.StopInfo
		data struct {
			StandardMoney float64 `bson:"standard_money"`
			 }
		resp = &response.JsonData{}
		err error
	)
	uid = c.GetString("uid")
	if !bson.IsObjectIdHex(uid) {
		resp.ExecFail(c, "用户ID异常")
		return
	}

	err = c.ShouldBindJSON(&carInfo)
	if err != nil {
		resp.ExecFail(c, "解析前端请求失败")
		return
	}

	//1. 查询车辆入场信息
	match = bson.M{"car_no": carInfo.CarNo, "place_no": carInfo.PlaceNo, "car_type": carInfo.CarType}
	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Find(match).One(&stopInfo)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "查询车辆停车信息失败")
		return
	}

	stopInfo.EndTime = time.Now().Unix()
	stopInfo.SumTime = stopInfo.EndTime - stopInfo.StartTime
	//2. 更新车辆入场信息
	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Update(match, bson.M{"$set": bson.M{"end_time": stopInfo.EndTime, "sum_time": stopInfo.SumTime}})
	})
	if err != nil {
		resp.ExecFail(c, "更新车辆停车信息失败")
		return
	}

	//3. 更新车位状态
	err = db.Mgo.Collection("Park", func(c *mgo.Collection) error {
		return c.Update(bson.M{"number": carInfo.PlaceNo, "car_type": carInfo.CarType}, bson.M{"$set": bson.M{
			"status": 1,
		}})
	})
	if err != nil {
		resp.ExecFail(c, "更新车位状态失败")
		return
	}

	//4. 根据车位类型计算费用 Billing standard
	match = bson.M{"car_type": carInfo.CarType}
	err = db.Mgo.Collection("BillStandard", func(c *mgo.Collection) error {
		return c.Find(match).One(&data)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "查询计费标准失败")
		return
	}
	if data.StandardMoney <= 0 {
		data.StandardMoney = 8
	}

	carStopInfo.CarNo = carInfo.CarNo
	carStopInfo.SumTime = time.Now().Unix() - stopInfo.StartTime
	if carStopInfo.SumTime % 60 > 30 {	//大于半小时按一小时算
		carStopInfo.Money = (float64(carStopInfo.SumTime) / 60.0 + 1) * data.StandardMoney
	} else if carStopInfo.SumTime % 60 > 0 && carStopInfo.SumTime < 30 {	//半小时内按半小时算
		carStopInfo.Money = (float64(carStopInfo.SumTime) / 60.0 + 0.5) * data.StandardMoney
	} else {
		carStopInfo.Money = (float64(carStopInfo.SumTime) / 60.0) * data.StandardMoney
	}
	resp.ExecSucc(c, carStopInfo)
}
