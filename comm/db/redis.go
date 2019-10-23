package db

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"log"
	"parking/comm/config"
	"time"
)

var Redis redisConfig

type redisConfig struct {
	MasterNewPool *redis.Pool
	SlaveNewPool  *redis.Pool
}

func init() {
	if err := config.Init(""); err != nil {
		log.Fatal("初始化配置文件失败", err)
	}
	Redis.MasterPool()
	Redis.SlavePool()
}

//redis write pool
func (r *redisConfig) MasterPool() {
	//var (
	//	err  error
	//	conn redis.Conn
	//)
	maxactive := viper.GetInt("redis.maxActive")
	wait := viper.GetBool("redis.wait")
	masterhost := viper.GetString("redis.mhost")
	masterpwd := viper.GetString("redis.mpwd")
	r.MasterNewPool = &redis.Pool{
		MaxActive:   maxactive,
		IdleTimeout: 300 * time.Second,
		Wait:        wait,
		//Dial: func() (redis.Conn, error) {
		//	if conn, err = redis.Dial("tcp", masterhost); err != nil {
		//		_ = conn.Close()
		//		log.InitLogger(log.ERROR).Error(err.Error(),zap.String(log.ERROR,errmsg.Redis_Master_ConnectErr))
		//		return nil, err
		//	}
		//	if _, err = conn.Do("auth", masterpwd); err != nil {
		//		_ = conn.Close()
		//		log.InitLogger(log.ERROR).Error(err.Error(),zap.String(log.ERROR,errmsg.Redis_Master_AuthErr))
		//		return nil, err
		//	}
		//	return conn, err
		//},
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", masterhost, redis.DialPassword(masterpwd))
			if nil != err {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

//redis read pool
func (r *redisConfig) SlavePool() {
	//var (
	//	err  error
	//	conn redis.Conn
	//)
	maxactive := viper.GetInt("redis.maxActive")
	wait := viper.GetBool("redis.wait")
	slavehost := viper.GetString("redis.shost")
	slavepwd := viper.GetString("redis.spwd")
	r.SlaveNewPool = &redis.Pool{
		MaxActive:   maxactive,
		IdleTimeout: 300 * time.Second,
		Wait:        wait,
		//Dial: func() (redis.Conn, error) {
		//	if conn, err = redis.Dial("tcp", slavehost,redis.DialPassword(slavepwd)); err != nil {
		//		_ = conn.Close()
		//		log.InitLogger(log.ERROR).Error(err.Error(),zap.String(log.ERROR,errmsg.Redis_Slave_ConnectErr))
		//		return nil, err
		//	}
		//	if _, err = conn.Do("auth", slavepwd); err != nil {
		//		_ = conn.Close()
		//		log.InitLogger(log.ERROR).Error(err.Error(),zap.String(log.ERROR,errmsg.Redis_Slave_AuthErr))
		//		return nil, err
		//	}
		//	return conn, err
		//},
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", slavehost, redis.DialPassword(slavepwd))
			if nil != err {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

const (
	SELECT           = "select"
	SET              = "set"
	GET              = "get"
	EX               = "ex"
	DEL              = "del"
	DBSIZE           = "dbsize"
	KEYS             = "keys"
	EXISTS           = "EXISTS"
	EXEC             = "EXEC"
	MULTI            = "multi"
	INCR             = "incr"
	DECR             = "decr"
	HSET             = "hset"
	HMSET            = "hmset"
	HGETALL          = "hgetall"
	HEXISTS          = "hexists"
	HDEL             = "hdel"
	SADD             = "sadd"
	ZADD             = "zadd"
	ZRANGE           = "zrange"
	ZCARD            = "zcard"
	HINCRBY          = "hincrby"
	SMEMBERS         = "smembers" // 返回集合中所有元素
	SISMEMBER        = "sismember" // 判断元素是否存在集合中
	ZRANGEBYSCORE    = "zrangebyscore"
	ZREMRANGEBYSCORE = "zremrangebyscore"
)

func errCheck(ms string, err error) {
	if err != nil {
		fmt.Printf("抱歉,遇到错误: %s.\r\n", ms, err)
		//os.Exit(-1)
	}
}

// set,json的struct
// data：结构对象
// 存切片：["aaa","bbb","ccc","ddd","sss"]
// 存map：{"1":"aaa","2":"bbb","3":"ccc","4":"ddd","5":"eee"}
func (r *redisConfig) SetJsonByStruct(key, exp string, data interface{}, num int) {
	var (
		err  error
		conn redis.Conn
		by   []byte
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	_, err = conn.Do(SELECT, num)

	by, err = json.Marshal(data)
	errCheck("json转换", err)

	if exp == "0" {
		_, err = conn.Do(SET, key, by)
		errCheck("保存时间", err)
	}
	_, err = conn.Do(SET, key, by, EX, exp)
}

// 读取结构类型的json串
func (r *redisConfig) GetJsonByStruct(key string, num int) []byte {
	var conn redis.Conn
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(GET, key)
	conn.Flush()
	conn.Receive()
	b, err := redis.Bytes(conn.Receive())
	errCheck("获取json值", err)
	return b
}

// 读取int类型的key-value
func (r *redisConfig) GetIntValue(key string, num int) int {
	var (
		err  error
		n    int
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(GET, key)
	conn.Flush()
	conn.Receive()
	n, err = redis.Int(conn.Receive())
	errCheck("获取int值", err)
	return n
}

// 读取int64类型的key-value
func (r *redisConfig) GetInt64Value(key string, num int) int64 {
	var (
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(GET, key)
	conn.Flush()
	conn.Receive()
	i, err := redis.Int64(conn.Receive())
	errCheck("获取int64值", err)
	return i
}

// 读取string类型的key-value
func (r *redisConfig) GetStringValue(key string, num int) string {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(GET, key)
	conn.Flush()
	conn.Receive()
	s, err := redis.String(conn.Receive())
	errCheck("获取string值", err) //
	return s
}

// 获取短信验证码
func (r *redisConfig) GetSMSCode(key string, num int) string {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(GET, key)
	conn.Flush()
	conn.Receive()
	if code, err := redis.String(conn.Receive()); err != nil {
		if code == "" && err.Error() == "redigo: nil returned" {
			return ""
		}
		return ""
	} else {
		return code
	}
}

// 读取[]string类型的key-value
func (r *redisConfig) GetStringsValue(key string, num int) []string {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(GET, key)
	conn.Flush()
	conn.Receive()
	s, err := redis.Strings(conn.Receive())
	errCheck("获取[]string值", err)
	return s
}

// 读取map[string]string类型的key-value
func (r *redisConfig) GetMapStringsValue(key string, num int) map[string]string {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(GET, key)
	conn.Flush()
	conn.Receive()
	sm, err := redis.StringMap(conn.Receive())
	errCheck("获取stringMap值", err)
	return sm
}

//----------------- key ------------------

// 读取[]interface类型的key-value
func (r *redisConfig) GetValues(key string, num int) []interface{} {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(GET, key)
	conn.Flush()
	conn.Receive()
	v, err := redis.Values(conn.Receive())
	errCheck("获取[]interface值", err)
	return v
}

// 查找指定模式的key,模糊匹配*user,user*,*user*,所有就传*
func (r *redisConfig) GetKeys(key string, num int) []string {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(KEYS, key)
	conn.Flush()
	conn.Receive()
	v, err := redis.Strings(conn.Receive())
	errCheck("搜索指定关键词的key", err)
	return v
}

// 添加key-value
func (r *redisConfig) SetKey(key, data string, num int) bool {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(MULTI)
	conn.Send(SELECT, num)
	conn.Send(SET, key, data)
	//conn.Flush()
	//conn.Receive()
	//conn.Receive()
	_, err := conn.Do(EXEC)
	if err != nil {
		return false
	}
	return true
}

// 添加key-value,并设置过期时间
func (r *redisConfig) SetKeyEXP(key, data, exp string, num int) bool {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(MULTI)
	conn.Send(SELECT, num)
	conn.Send(SET, key, data, "EX", exp)
	_, err := conn.Do(EXEC)
	if err != nil {
		return false
	}
	return true
}

// 删除key
func (r *redisConfig) DeleteKey(key string, num int) bool {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(MULTI)
	conn.Send(SELECT, num)
	conn.Send(DEL, key)
	_, err := conn.Do(EXEC)
	if err != nil {
		fmt.Printf("抱歉,遇到错误: %s.\r\n", err)
		return false
	}
	return true
}

// 获取当前库的key数量
// keys *这种数据量小还可以，大的时候可以直接搞死生产环境。
// dbsize和keys *统计的key数可能是不一样的，如果没记错的话，keys *统计的是当前db有效的key，而dbsize统计的是所有未被销毁的key
func (r *redisConfig) GetKeyCount(num int) int {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(DBSIZE)
	conn.Flush()
	conn.Receive()
	i, err := redis.Int(conn.Receive())
	errCheck("获取key数", err)
	return i
}

// 获取所有key
func (r *redisConfig) GetAllKey(num int) []string {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(KEYS, "*")
	conn.Flush()
	conn.Receive()
	s, err := redis.Strings(conn.Receive())
	errCheck("获取[]string数", err)
	return s
}

// 判断指定key是否存在
func (r *redisConfig) ExistKey(key string, num int) bool {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(EXISTS, key)
	conn.Flush()
	conn.Receive()
	b, err := redis.Bool(conn.Receive())
	errCheck("key是否存在", err)
	return b
}

// 自增长
func (r *redisConfig) KeyINCR(key string, num int) bool {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(MULTI)
	conn.Send(INCR, key)
	_, err := conn.Do(EXEC)
	errCheck("自增长", err)
	return true
}

// 自减
func (r *redisConfig) KeyDECR(key string, num int) bool {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(MULTI)
	conn.Send(DECR, key)
	_, err := conn.Do(EXEC)
	errCheck("自增长", err)
	return true
}

// 自减,并返回
func (r *redisConfig) GetKeyDECR(key string, num int) int64 {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(MULTI)
	conn.Send(DECR, key)
	i, err := conn.Do(EXEC)
	errCheck("自减,并返回", err)
	return i.(int64)
}

// 自增长,并返回
func (r *redisConfig) GetKeyINCR(key string, num int) int64 {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	i, err := redis.Int64(conn.Do(INCR, key))
	errCheck("自增长并返回", err)
	return i

}

// -------------- hash ----------------

// 删除hash中指定field
func (r *redisConfig) DeleteFieldFromHash(key, field string, num int) bool {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(HDEL, key, field)
	_, err := conn.Do(EXEC)
	errCheck("删除hash中指定field", err)
	return true
}

// 判断hash中字段是否存在
func (r *redisConfig) ExistFieldFromHash(key, field string, num int) bool {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(HEXISTS, key, field)
	conn.Flush()
	conn.Receive()
	v, err := redis.Bool(conn.Receive())
	errCheck("判断hash中字段是否存在", err)
	return v
}

// struct to hash 结构体保存为哈希
func (r *redisConfig) SetStructToHash(key string, fieldValue interface{}, num int) bool {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(HMSET, redis.Args{key}.AddFlat(fieldValue)...)
	conn.Do(EXEC)
	errCheck("map存储为hash", err)
	return true
}

// hash to struct 从哈希获取结构体
/*
	调用例子：
	var u = new(User)
	v,_ := db.Redis.GetStructFromHash("ppp",13)
	err:=redis.ScanStruct(v,u)
	fmt.Println(u.Job,u.Age,u.Name,err)
*/
func (r *redisConfig) GetStructFromHash(key string, num int) ([]interface{}, error) {
	var (
		err error
		v   []interface{}
	)
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(HGETALL, key)
	err = conn.Flush()
	conn.Receive()
	v, err = redis.Values((conn.Receive()))
	if err != nil {
		return nil, err
	}
	return v, nil
}

// map存储为hash
func (r *redisConfig) SetMapToHash(key string, fieldValue map[string]interface{}, num int) bool {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(HMSET, redis.Args{}.Add(key).AddFlat(fieldValue)...)
	conn.Do(EXEC)
	errCheck("map存储为hash", err)
	return true
}

// hash返回map[string]string
func (r *redisConfig) GetHashMapString(key string, num int) map[string]string {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(HGETALL, key)
	conn.Flush()
	conn.Receive()
	sm, err := redis.StringMap(conn.Receive())
	errCheck("hash返回map[string]string", err)
	return sm
}

// hash返回map[string]int
func (r *redisConfig) GetHashMapInt(key string, num int) map[string]int {
	conn := r.SlaveNewPool.Get()
	conn.Send(SELECT, num)
	conn.Send(HGETALL, key)
	conn.Flush()
	conn.Receive()
	sm, err := redis.IntMap(conn.Receive())
	errCheck("hash返回map[string]int", err)
	return sm
}

// hash返回map[string]int64
func (r *redisConfig) GetHashMapInt64(key string, num int) map[string]int64 {
	conn := r.SlaveNewPool.Get()
	conn.Send(SELECT, num)
	conn.Send(HGETALL, key)
	conn.Flush()
	conn.Receive()
	sm, err := redis.Int64Map(conn.Receive())
	errCheck("hash返回map[string]int64", err)
	return sm
}

//hash返回map[string]interface{}
func (r *redisConfig) GetHashMapInterface(key string, num int) map[string]interface{} {
	var (
		dKey string
		value interface{}
		m = make(map[string]interface{})
		err error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	conn.Send(SELECT, num)
	conn.Send(HGETALL, key)
	conn.Flush()
	conn.Receive()
	data, err := redis.Values(conn.Receive())
	errCheck("hans返回map[string]interface", err)
	for k, v := range data {
		if k % 2 == 0 {
			switch v.(type) {
			case []byte:
				dKey = string(v.([]byte))
			}
		} else {
			switch v.(type) {
			case []byte:
				value = string(v.([]byte))
			}
			m[dKey] = value
		}
	}
	return m
}

// 写入结构
// interface{}不需要&原本就是一个指针
func (r *redisConfig) SetHashStruct(key string, data interface{}, num int) (interface{}, error) {
	conn := r.SlaveNewPool.Get()
	conn.Do("select", num)
	defer conn.Close()
	return conn.Do(HMSET, redis.Args{}.Add(key).AddFlat(data)...)
}

// 设置hash值
func (r *redisConfig) SetHashOneKey(key, subKey string, data interface{}, num int) bool {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(HSET, key, subKey, data)
	conn.Do(EXEC)
	errCheck("写入hash值", err)
	return true
}

// hash字段增减(判断不能为负数)
func (r *redisConfig) SetHashHIncrBy(key, subKey string, i, num int) bool {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	conn.Send(HGETALL, key)
	conn.Flush()
	conn.Receive()
	sm, err := redis.IntMap(conn.Receive())
	if sm[subKey] > 0 {
		err = conn.Send(HINCRBY, key, subKey, i)
		conn.Do(EXEC)
	}
	errCheck("hash值自增数量", err)
	return true
}

// hash字段减
func (r *redisConfig) SetHashDecrby(key, subKey string, num int) bool {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send("hdecrby", key, subKey)
	conn.Do(EXEC)
	errCheck("hash值自减数量", err)
	return true
}

// ------------------------ 集合 --------------------------------------
// 写入集合
func (r *redisConfig) SetSADD(key string, data string, num int) bool {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(SADD, key, data)
	conn.Do(EXEC)
	errCheck("写入集合", err)
	return true
}

//写入切片到集合
func (r *redisConfig) SetSliceSADD(key string, data []interface{}, num, t int) bool {
	var (
		err error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	for _, v := range data {
		err = conn.Send(SADD, key, v)
	}
	//err = conn.Send(SADD, key, data)
	err = conn.Send("EXPIRE", t)
	conn.Do(EXEC)
	errCheck("写入切片到集合", err)
	return true
}

// 判断集合中是否存在该值
func (r *redisConfig) GetSISMEMBER(key, data string, num int) int64 {
	var (
		err  error
		v int64
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(SISMEMBER, key, data)
	err = conn.Flush()
	conn.Receive()
	v, err = redis.Int64(conn.Receive())
	errCheck("读取集合中单个元素值", err)
	return v
}

// 获取集合中所有元素
func (r *redisConfig) GetAllSub(key string, num int) []string {
	var (
		err error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	v, err := redis.Strings(conn.Do("SMEMBERS", key))
	errCheck("读取集合所有元素", err)
	return v
}

// -------------- 有序集合 ---------------

// 写入有序集合
func (r *redisConfig) SetSort(key string, index int, data string, num int) bool {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(ZADD, key, index, data)
	conn.Do(EXEC)
	errCheck("写入有序集合", err)
	return true
}

// 读取有序集合中单个元素值
func (r *redisConfig) GetSortValue(key string, index int, num int) string {
	var (
		err  error
		str  = ""
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(ZRANGEBYSCORE, key, index, index)
	err = conn.Flush()
	conn.Receive()
	v, err := redis.Strings(conn.Receive())
	errCheck("读取有序集合中单个元素值", err)
	for _, s := range v {
		str = s
	}
	return str
}

// 读取有序集合中所有元素(根据索引范围获取元素)
func (r *redisConfig) GetSortAllValue(key string, minIndex, MiaIndex int, num int) []string {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	conn.Send(SELECT, num)
	conn.Send(ZRANGE, key, minIndex, MiaIndex)
	conn.Flush()
	conn.Receive()
	v, err := redis.Strings(conn.Receive())
	errCheck("读取有序集合中所有元素(根据索引范围获取元素)", err)
	return v
}

// 获取有序集合中元素的数量
func (r *redisConfig) GetSortCount(key string, num int) int {
	var (
		err error
		n   int
	)
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(ZCARD, key)
	err = conn.Flush()
	conn.Receive()
	n, err = redis.Int(conn.Receive())
	errCheck("获取有序集合中元素的数量", err)
	return n
}

// 删除有序集合中某条或者范围内的记录，根据分数定位
func (r *redisConfig) DeleteSortValue(key string, min, max int, num int8) interface{} {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	conn.Send(ZREMRANGEBYSCORE, key, min, max)
	conn.Do(EXEC)
	errCheck("删除有序集合中某条或者范围内的记录，根据分数定位", err)
	return true
}

// 写入集合
func (r *redisConfig) ZSetSADDInt(key string, score, value int, num int) bool {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(ZADD, key, score, value)
	conn.Do(EXEC)
	errCheck("写入集合", err)
	return true
}

func (r *redisConfig)ZSetExietSub(key string, index int, num int ) bool {
	conn := r.SlaveNewPool.Get()
	defer conn.Close()
	return true
}

func (r *redisConfig) SetMapIntToHash(key string, fieldValue map[string]int, num int) bool {
	var (
		err  error
		conn redis.Conn
	)
	conn = r.SlaveNewPool.Get()
	defer conn.Close()
	err = conn.Send(SELECT, num)
	err = conn.Send(HMSET, redis.Args{}.Add(key).AddFlat(fieldValue)...)
	conn.Do(EXEC)
	errCheck("map存储为hash", err)
	return true
}

