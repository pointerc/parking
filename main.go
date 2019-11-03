package main

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"parking/comm/db"
	_ "parking/goapp/handle/space_handle"
	"parking/goapp/model/space_model"
	"parking/goapp/service"
	"syscall"
	"unsafe"
)

var (
	max = viper.GetInt("parking.max")
	min = viper.GetInt("parking.min")
)

const (
	IpcCreate = 00001000
)

func init() {
	var (
		//isInit = viper.GetInt("init")
		ParkSpace space_model.ParkingSpace
		MaxPark = make([]interface{}, max)
		MinPark = make([]interface{}, min)
		err error
	)
	//使用共享内存，不删除该共享内存，保证该数据不管是否重启服务，只初始化一次
	shmid, _, err := syscall.Syscall(syscall.SYS_SHMGET, 1024, 16, IpcCreate|0600)
	shmaddr, _, err := syscall.Syscall(syscall.SYS_SHMAT, shmid, 0, 0)
	//fmt.Println(*(*int)(unsafe.Pointer(uintptr(shmaddr))))
	if *(*int)(unsafe.Pointer(uintptr(shmaddr))) == 0 {
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
			log.Fatal("创建大型车位失败")
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
			log.Fatal("创建小型车位失败")
			return
		}
		*(*int)(unsafe.Pointer(uintptr(shmaddr))) = 1
	}
}

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
		//fmt.Println("method:", method)
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
