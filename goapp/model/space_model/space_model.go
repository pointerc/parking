package space_model

import "gopkg.in/mgo.v2/bson"

type ParkingSpace struct {
	ID       bson.ObjectId `json:"id" bson:"_id"`
	Number   int           `json:"number" bson:"number"`       //车位号
	Status   int           `json:"status" bson:"status"`       //车位状态 1:未被占用 2:已被占用
	ParkType int           `json:"park_type" bson:"park_type"` //车位类型 1:小型车位 2:大型车位
	Flag     int           `json:"flag" bson:"flag"`           //车位使用状态 1:使用中 2:已删除
}
