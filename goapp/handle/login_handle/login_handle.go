package login_handle

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
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
		resp.ExecFail(c, "解析前端请求失败")
		return
	}
	err = db.Mgo.Collection("User", func(c *mgo.Collection) error {
		return c.Find(bson.M{"username": userInfo.NickName}).One(&data)
	})
	if err != nil && err.Error() != "not found" {
		resp.ExecFail(c, "系统异常")
		return
	}
	if data.ID.Hex() != "" {
		resp.ExecFail(c, "该用户名已存在")
		return
	}
	m := md5.New()
	v, _ := json.Marshal(userInfo)
	m.Write(v)
	token := hex.EncodeToString(m.Sum(nil))
	resp.RegisterSucc(c, token)
	go func() {
		//将token写入redis
	}()
}