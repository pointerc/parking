package login_model

import (
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID bson.ObjectId `json:"uid" bson:"_id"`	//用户id
	NickName string `json:"nick_name" bson:"nick_name"`
	PassWord string `json:"pass_word" bson:"pass_word"`
}


//校验用户名密码
func CheckUserPasswd(user *User) {
	var (

	)

}
