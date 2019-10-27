package car_model

import "gopkg.in/mgo.v2/bson"

type CarInfo struct {
	ID         bson.ObjectId `json:"id" bson:"_id"`
	SpaceId    bson.ObjectId `json:"space_id" bson:"space_id"`                   //车位id
	CarNo      string        `json:"carNo" bson:"car_no" binding:"required"`     //卡号
	PlaceNo    int        `json:"placeNo" bson:"place_no" binding:"required"` //车位号
	CarType    int           `json:"carType" bson:"car_type" binding:"required"` //车类型
	CreateTime string        `json:"create_time" bson:"create_time"`             //创建时间
	StartTime  int64         `json:"start_time" bson:"start_time"`               //车入场时间
	EndTime    int64         `json:"end_time" bson:"end_time"`                   //车出场时间
	SumTime    int64         `json:"sumTime" bson:"sum_time"`                    //停车时长
}

type StopInfo struct {
	CarNo   string  `json:"carNo" bson:"car_no"`     //卡号
	SumTime int64   `json:"sumTime" bson:"sum_time"` //停车时间
	Money   float64 `json:"money" bson:"money"`      //应收费用
}
