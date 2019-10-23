package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type JsonData struct {
	State int `json:"state"`	//0:成功 1:失败
	Code int `json:"code"`		//
	Data interface{} `json:"data"`	//返回信息
	Msg string `json:"msg"`		//返回信息
	Token string `json:"token"`	//用户token
}

func (j *JsonData)ExecSucc(c *gin.Context, data interface{}) {
	var result JsonData
	result.State = 0
	result.Code = 200
	result.Data = data
	result.Msg = ""
	result.Token = ""
	c.JSON(http.StatusOK, result)
}

func (j *JsonData)ExecFail(c *gin.Context, msg string) {
	var result JsonData
	result.State = 1
	result.Code = http.StatusNoContent
	result.Data = struct {}{}
	result.Msg = msg
	result.Token = ""
	c.JSON(http.StatusNoContent, result)
}

func (j *JsonData)RegisterSucc(c *gin.Context, token string) {
	var result JsonData
	result.State = 0
	result.Code = http.StatusOK
	result.Data = []struct{}{}
	result.Msg = "注册成功"
	result.Token = token
	c.JSON(http.StatusOK, result)
}
