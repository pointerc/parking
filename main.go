package main

import (
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"net/http"
	_ "parking/goapp/handle/space_handle"
	"parking/goapp/service"
)

func main() {
	Start()
}

func Start() {
	gin.SetMode("debug") //设置gin的模式
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression)) //设置请求数据压缩
	router.MaxMultipartMemory = 8 << 20

	// Recovery 中间件会 recover 任何 panic。如果有 panic 的话，会写入 500。
	router.Use(gin.Recovery())

	//gin解决跨域问题
	router.Use(Cors())

	Addrouter(router)

	router.Run("0.0.0.0:8090")
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
		fmt.Println("method:", method)
		//允许所有OPTIONS请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		//忽略该路由请求
		if c.Request.URL.Path == "/favicon.ico" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		//处理请求
		c.Next()
	}
}

func Addrouter(router *gin.Engine) {
	var (
		ILogin service.ILogin
		login  = new(service.Login)

		ISpace service.ISpace
		space = new(service.Space)

		ICar service.ICar
		car = new(service.Car)
	)

	//注册登录
	ILogin = login
	ILogin.LoginRouter(router)

	//创建车位
	ISpace = space
	ISpace.SpaceRoter(router)

	//车辆出入场
	ICar = car
	ICar.CarStop(router)

	//收费标准

	//

	//

	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "没有这个路由")
	})
}
