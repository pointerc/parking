package login_handle

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"parking/comm/db"
	"parking/comm/response"
	"parking/goapp/model/login_model"
)

type Login struct{}

func (l *Login)UserLogin(c *gin.Context) {
	var (
		UserLogin login_model.User
		data = make(map[string]interface{})
		err error
	)

	//获取前端请求用户名与密码
	err = c.ShouldBindJSON(&UserLogin)
	if err != nil {
		data["msg"] = "解析前端请求失败"
		c.JSON(http.StatusBadRequest, data)
		return
	}
}

func (l *Login) Register(c *gin.Context) {
	var (
		userInfo login_model.User
		data login_model.User
		resp = &response.JsonData{}
		err error
	)
	err = c.ShouldBindJSON(&userInfo)
	if err != nil {
		fmt.Println("解析前端请求失败:", err)
		resp.ExecFail(c, "解析前端请求失败")
		return
	}
	fmt.Println("前端请求：", userInfo)
	err = db.Mgo.Collection("User", func(c *mgo.Collection) error {
		return c.Find(bson.M{"nick_name": userInfo.NickName}).One(&data)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "系统异常")
		return
	}
	if data.ID.Hex() != "" {
		resp.ExecFail(c, "该用户名已存在")
		return
	}
	userInfo.ID = bson.NewObjectId()
	fmt.Println("insert mongodb:", userInfo)
	db.Mgo.Collection("User", func(c *mgo.Collection) error {
		return c.Insert(userInfo)
	})
	tokenAES := struct {
		Uid bson.ObjectId `json:"uid"`
	}{Uid: userInfo.ID}
	v, _ := json.Marshal(tokenAES)
	token := base64.StdEncoding.EncodeToString(v)
	resp.RegisterSucc(c, token)
}