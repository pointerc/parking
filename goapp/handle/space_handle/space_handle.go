package space_handle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"parking/comm/db"
	"parking/comm/response"
	"parking/goapp/model/space_model"
	"sync"
)

type Space struct {}

var (
	max = viper.GetInt("parking.max")
	min = viper.GetInt("parking.min")
)

func (s *Space) SpaceInit(c *gin.Context) {
	var (
		uid string
		once sync.Once
		ParkSpace space_model.ParkingSpace
		MaxPark = make([]interface{}, max)
		MinPark = make([]interface{}, min)
		resp = &response.JsonData{}
		err error
	)
	once.Do(func() {
		uid = c.GetString("uid")
		if !bson.IsObjectIdHex(uid) {
			resp.ExecFail(c, "用户ID异常")
			return
		}

		for i := 0; i < max; i++ {
			ParkSpace = space_model.ParkingSpace{
				ID: bson.NewObjectId(),
				Number: i + 1,
				Status: 1,
				ParkType: 2,
				Flag: 1,
			}
			MaxPark = append(MaxPark, ParkSpace)
		}
		err = db.Mgo.Collection("ParkSpace", func(c *mgo.Collection) error {
			return c.Insert(MaxPark...)
		})
		if err != nil {
			resp.ExecFail(c, "系统异常")
			return
		}

		for i := 0; i < min; i++ {
			ParkSpace = space_model.ParkingSpace{
				ID: bson.NewObjectId(),
				Number: i + max + 1,
				Status: 1,
				ParkType: 1,
				Flag: 1,
			}
			MinPark = append(MinPark, ParkSpace)
		}
		err = db.Mgo.Collection("ParkSpace", func(c *mgo.Collection) error {
			return c.Insert(MinPark...)
		})
		if err != nil {
			resp.ExecFail(c, "系统异常")
			return
		}
		resp.Succ(c)
	})
}

//剩余车位数
func (s *Space) FreeSpace(c *gin.Context) {
	var (
		count int
		uid string
		match bson.M
		resp = &response.JsonData{}
		data = make(map[string]interface{})
		err error
	)
	uid = c.GetString("uid")
	fmt.Println("uid:", uid)
	if !bson.IsObjectIdHex(uid) {
		resp.ExecFail(c, "用户ID异常")
		return
	}

	//查询大型车车位剩余数量
	match = bson.M{"park_type": 2, "status": 1}
	err = db.Mgo.Collection("ParkSpace", func(c *mgo.Collection) error {
		count, err = c.Find(match).Count()
		return err
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "查询大型车辆车位剩余数失败")
		return
	}
	data["max"] = count

	//查询小型车车位剩余数量
	match = bson.M{"park_type": 1, "status": 1}
	err = db.Mgo.Collection("ParkSpace", func(c *mgo.Collection) error {
		count, err = c.Find(match).Count()
		return err
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "查询小型车辆车位剩余数失败")
		return
	}
	data["min"] = count
	resp.ExecSucc(c, data)
}
