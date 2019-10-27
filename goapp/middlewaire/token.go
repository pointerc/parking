package middlewaire

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"parking/comm/response"
)

type Token struct {
	UID string
	Token string `json:"token"`
}

func (t *Token)CheckToken() gin.HandlerFunc{
	return func(c *gin.Context) {
		var (
			token string
			data = make(map[string]interface{})
			resp = &response.JsonData{}
			err error
		)
		token = c.Request.Header.Get("token")
		if token == "" {
			resp.ExecFail(c, "token无效")
			c.Abort()
			return
		}
		b, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			resp.ExecFail(c, "解析token失败")
			c.Abort()
			return
		}
		err = json.Unmarshal(b, &data)
		if err != nil {
			resp.ExecFail(c, "解析token失败")
			c.Abort()
			return
		}
		c.Set("uid", data["uid"])
		c.Next()
		return
	}
}
