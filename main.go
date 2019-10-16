package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	//router := gin.Default()
	//router.GET("/", func(c *gin.Context) {
	//	c.String(200, "test")
	//})
	//router.Run(":8080")
}

func Start() {
	gin.SetMode("debug")	//设置gin的模式
	//router := gin.Default()
	router := gin.Default()
	//router.Use(gzip.Gzip(gzip.DefaultCompression)) //设置请求数据压缩
	router.MaxMultipartMemory = 8 << 20

	// Recovery 中间件会 recover 任何 panic。如果有 panic 的话，会写入 500。
	router.Use(gin.Recovery())

	//gin解决跨域问题
	router.Use(Cors())

	Addrouter(router)

	router.Run(":8090")
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Content-Type", "application/json")

		//允许所有OPTIONS请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		//处理请求
		c.Next()
	}
}

func Addrouter(router *gin.Engine) {

}