package car_handle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"parking/comm/db"
	"parking/comm/response"
	"parking/goapp/model/car_model"
	"strconv"
	"time"
)

type Car struct{}

//车辆入场
func (car *Car) VehileAdmission(c *gin.Context) {
	var (
		number, park_type int
		uid               string
		req               struct {
							  CarNo   string `json:"carNo" bson:"car_no"`
							  CarType string `json:"carType" bson:"park_type"`
							  PlaceNo string `json:"placeNo" bson:"number"`
						  }
		carInfo    car_model.CarInfo
		oldCarInfo car_model.CarInfo
		match      bson.M
		ParkInfo   struct {
					   ID     bson.ObjectId `bson:"_id"`
					   Status int           `bson:"status"`
				   }
		resp = &response.JsonData{}
		err  error
	)

	uid = c.GetString("uid")
	if !bson.IsObjectIdHex(uid) {
		resp.ExecFail(c, "用户ID异常")
		return
	}

	err = c.ShouldBindJSON(&req)
	if err != nil {
		resp.ExecFail(c, "解析前端请求失败")
		return
	}

	//判断卡号是否存在
	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Find(bson.M{"car_no": carInfo.CarNo, "end_time": 0}).One(&oldCarInfo)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "查询用户是否已存在停车信息失败")
		return
	}
	if oldCarInfo.ID.Hex() != "" {
		resp.ExecFail(c, "您已有车辆停入")
		return
	}
	number, _ = strconv.Atoi(req.PlaceNo)
	park_type, _ = strconv.Atoi(req.CarType)
	//根据车位号，车类型查询车位id
	match = bson.M{"number": number, "park_type": park_type, "status": 1}
	fmt.Println(match)
	err = db.Mgo.Collection("ParkSpace", func(c *mgo.Collection) error {
		return c.Find(match).One(&ParkInfo)
	})
	fmt.Println("error:", err)
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "系统异常")
		return
	}
	fmt.Println(ParkInfo)
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
	carInfo.CarNo = req.CarNo
	carInfo.CarType = park_type
	carInfo.PlaceNo = number
	carInfo.SpaceId = ParkInfo.ID
	carInfo.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	carInfo.StartTime = time.Now().Unix()
	carInfo.EndTime = 0
	carInfo.Money = 0
	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Insert(carInfo)
	})

	//更新车位状态
	if err != nil {
		resp.ExecFail(c, "车辆入场失败")
		return
	}

	err = db.Mgo.Collection("ParkSpace", func(c *mgo.Collection) error {
		return c.Update(bson.M{"_id": ParkInfo.ID}, bson.M{"$set": bson.M{"status": 2}})
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
		uid           string
		match         bson.M
		carInfo       car_model.CarInfo
		stopInfo      car_model.CarInfo
		carStopInfo   car_model.StopInfo
		StandardMoney float64
		resp          = &response.JsonData{}
		err           error
	)
	uid = c.GetString("uid")
	if !bson.IsObjectIdHex(uid) {
		resp.ExecFail(c, "用户ID异常")
		return
	}

	err = c.ShouldBindJSON(&carInfo)
	if err != nil {
		fmt.Println("error:", err)
		resp.ExecFail(c, "解析前端请求失败")
		return
	}

	//1. 查询车辆入场信息
	match = bson.M{"car_no": carInfo.CarNo, "place_no": carInfo.PlaceNo, "car_type": carInfo.CarType, "end_time": 0}
	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Find(match).One(&stopInfo)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "查询车辆停车信息失败")
		return
	}

	stopInfo.EndTime = time.Now().Unix()
	stopInfo.SumTime = stopInfo.EndTime - stopInfo.StartTime

	//4. 根据车位类型计算费用 Billing standard
	if carInfo.CarType == 1 {
		StandardMoney = viper.GetFloat64("park.min_price")
	} else if carInfo.CarType == 2 {
		StandardMoney = viper.GetFloat64("park.max_price")
	} else {
		StandardMoney = (viper.GetFloat64("park.min_price") + viper.GetFloat64("park.max_price")) / 2.0
	}

	carStopInfo.CarNo = carInfo.CarNo
	carStopInfo.SumTime = time.Now().Unix() - stopInfo.StartTime
	if carStopInfo.SumTime/60%60 > 30 { //大于半小时按一小时算
		carStopInfo.Money = float64(carStopInfo.SumTime/3600 + 1) * StandardMoney
	} else if carStopInfo.SumTime/60%60 > 0 && carStopInfo.SumTime < 30 { //半小时内按半小时算
		carStopInfo.Money = (float64(carStopInfo.SumTime)/3600.0 + 0.5) * StandardMoney
	} else {
		carStopInfo.Money = (float64(carStopInfo.SumTime) / 3600.0) * StandardMoney
	}
	carStopInfo.Money, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", carStopInfo.Money), 64)
	//2. 更新车辆入场信息
	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Update(match, bson.M{"$set": bson.M{
			"end_time": stopInfo.EndTime,
			"sum_time": stopInfo.SumTime,
			"out_time": time.Now().Format("2006-01-02 15:04:05"),
			"money":    carStopInfo.Money}})
	})
	if err != nil {
		resp.ExecFail(c, "更新车辆停车信息失败")
		return
	}

	//3. 更新车位状态
	err = db.Mgo.Collection("ParkSpace", func(c *mgo.Collection) error {
		return c.Update(bson.M{"number": carInfo.PlaceNo, "park_type": carInfo.CarType}, bson.M{"$set": bson.M{
			"status": 1,
		}})
	})
	if err != nil {
		resp.ExecFail(c, "更新车位状态失败")
		return
	}

	resp.ExecSucc(c, carStopInfo)
}

//查询停车场车辆信息
func (car *Car) GetParkInfo(c *gin.Context) {
	var (
		uid      string
		match    bson.M
		ParkInfo []struct {
			ID         bson.ObjectId `json:"id" bson:"_id"`
			CarNo      string        `json:"carNo" bson:"car_no"`
			PlaceNo    int           `json:"placeNo" bson:"place_no"`
			CarType    int           `json:"carType" bson:"car_type"`
			CreateTime string        `json:"create_time" bson:"create_time"` //创建时间
		}
		resp = &response.JsonData{}
		err  error
	)
	uid = c.GetString("uid")
	if !bson.IsObjectIdHex(uid) {
		resp.ExecFail(c, "用户ID不合法")
		return
	}

	//查询当前停车场停车信息
	match = bson.M{"start_time": bson.M{"$gt": 0}, "end_time": 0}
	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Find(match).Select(bson.M{"_id": 1, "car_no": 1, "place_no": 1, "car_type": 1, "create_time": 1}).All(&ParkInfo)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "查询当前停车场信息失败")
		return
	}
	if len(ParkInfo) == 0 {
		fmt.Println("未查询到数据")
		resp.ExecSucc(c, []struct{}{})
		return
	}
	resp.ExecSucc(c, ParkInfo)

}

//收费标准
func (car *Car) GetParkPrice(c *gin.Context) {
	var (
		data = make(map[string]interface{})
		resp = &response.JsonData{}
	)
	data["max_price"] = viper.GetInt("park.max_price")
	data["min_price"] = viper.GetInt("park.min_price")
	resp.ExecSucc(c, data)
}

//查询用户停车记录
func (car *Car) GetUserHistory(c *gin.Context) {
	var (
		uid     string
		history []car_model.CarInfo
		req     struct {
					CarNo string `json:"carNo" bson:"car_no"`
				}
		resp = &response.JsonData{}
		err  error
	)
	uid = c.GetString("uid")
	if !bson.IsObjectIdHex(uid) {
		resp.ExecFail(c, "用户ID不合法")
		return
	}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		resp.ExecFail(c, "解析前端请求失败")
		return
	}
	//carNo = c.Query("carNo")
	//if carNo == "" {
	//	resp.ExecFail(c, "卡号不能为空")
	//	return
	//}

	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Find(bson.M{"car_no": req.CarNo}).All(&history)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "查询用户历史停车信息失败")
		return
	}
	if len(history) == 0 {
		resp.ExecSucc(c, []struct{}{})
		return
	}
	//
	//data := make([]map[string]interface{}, len(history))
	//for i, v := range history {
	//	data[i] = make(map[string]interface{})
	//	data[i][""]
	//}

	resp.ExecSucc(c, history)
}

//查询用户当前停车信息
func (car *Car) GetNowParkInfo(c *gin.Context) {
	var (
		uid           string
		match         bson.M
		StandardMoney float64
		req           struct {
						  CarNo string `json:"carNo" bson:"car_no" binding:"required"`
					  }
		carInfo struct {
			ID        bson.ObjectId `json:"id" bson:"_id"`
			CarNo     string        `json:"carNo" bson:"car_no"`
			PlaceNo   int           `json:"placeNo" bson:"place_no"`
			CarType   int           `json:"carType" bson:"car_type"`
			StartTime int64         `json:"startTime" bson:"start_time"` //车入场时间
			SumTime   int64         `json:"sumTime" bson:"sum_time"`
			Money     float64       `json:"money" bson:"money"`
		}
		resp = &response.JsonData{}
		err  error
	)
	uid = c.GetString("uid")
	if !bson.IsObjectIdHex(uid) {
		resp.ExecFail(c, "用户ID不合法")
		return
	}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		resp.ExecFail(c, "解析前端请求失败")
		return
	}

	//根据卡号，查询用户当天停车信息
	match = bson.M{"car_no": req.CarNo, "end_time": 0}
	err = db.Mgo.Collection("Car", func(c *mgo.Collection) error {
		return c.Find(match).Select(bson.M{"_id": 1, "place_no": 1, "car_type": 1, "start_time": 1, "car_no": 1}).One(&carInfo)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "查询用户停车信息失败")
		return
	}
	if carInfo.ID.Hex() == "" {
		resp.ExecFail(c, "未查询到用户当前停车信息")
		return
	}

	if carInfo.CarType == 1 {
		StandardMoney = viper.GetFloat64("park.min_price")
	} else if carInfo.CarType == 2 {
		StandardMoney = viper.GetFloat64("park.max_price")
	} else {
		StandardMoney = (viper.GetFloat64("park.min_price") + viper.GetFloat64("park.max_price")) / 2.0
	}

	carInfo.SumTime = time.Now().Unix() - carInfo.StartTime
	//fmt.Println("sum_time", carInfo.SumTime % 60)
	if carInfo.SumTime/60%60 > 30 { //大于半小时按一小时算
		carInfo.Money = float64(carInfo.SumTime/3600 + 1) * StandardMoney
	} else if carInfo.SumTime/60%60 > 0 && carInfo.SumTime < 30 { //半小时内按半小时算
		carInfo.Money = float64(float64(carInfo.SumTime)/3600.0 + 0.5) * StandardMoney
	} else {
		carInfo.Money = (float64(carInfo.SumTime) / 3600.0) * StandardMoney
	}
	carInfo.Money, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", carInfo.Money), 64)
	//fmt.Println("carInfo:", carInfo)
	resp.ExecSucc(c, carInfo)
}
