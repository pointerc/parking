package db

import (
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"parking/comm/config"
)

var (
	Mgo mgoConfig
	url string
	db  string
)

type mgoConfig struct {
	NewPool *mgo.Session
}

func init() {
	if err := config.Init(""); err != nil {
		fmt.Println("初始化配置文件失败", err)
		panic(err)
	}
	url = viper.GetString("mongodb.url")
	db = viper.GetString("mongodb.db")
	//Mgo.Pool()
}

func (m *mgoConfig) Pool() *mgo.Session {
	if m.NewPool == nil {
		var err error
		m.NewPool, err = mgo.Dial(url)
		if err != nil {
			log.Fatal("创建mongo连接池失败", err)
			panic(err)
		}
		m.NewPool.SetMode(mgo.Eventual, true)
	}
	return m.NewPool.Clone()
}

//collection opt
func (m *mgoConfig) Collection(collection string, getCollection func(*mgo.Collection) error) error {
	session := Mgo.Pool()
	defer session.Close()
	coll := Mgo.NewPool.DB(db).C(collection)
	return getCollection(coll)
}

func (m *mgoConfig) BsonToObject(val interface{}, obj interface{}) error {
	data, err := bson.Marshal(val)
	if err != nil {
		return err
	}
	bson.Unmarshal(data, obj)
	return nil
}

// find one
func (m *mgoConfig) FindOne(TableName string, query, selector bson.M) (result bson.M, err error) {
	exop := func(c *mgo.Collection) error { return c.Find(query).Select(selector).One(&result) }
	err = m.Collection(TableName, exop)
	return result, err
}

// find more
func (m *mgoConfig) FindMore(TableName string, query, selector bson.M) (result []bson.M, err error) {
	exop := func(c *mgo.Collection) error {
		return c.Find(query).Select(selector).All(&result)
	}
	err = m.Collection(TableName, exop)
	return
}

// find more,sort,skip,limit
func (m *mgoConfig) FindMoreLimit(TableName string, query bson.M, sort string, fields bson.M, skip int, limit int) (result []bson.M, err error) {
	exop := func(c *mgo.Collection) error {
		if sort != "" {
			return c.Find(query).Sort(sort).Select(fields).Skip(skip).Limit(limit).All(&result)
		} else {
			return c.Find(query).Select(fields).Skip(skip).Limit(limit).All(&result)
		}
	}
	err = m.Collection(TableName, exop)
	return
}

//  update
func (m *mgoConfig) Update(collection string, query, selector, result interface{}) (interface{}, error) {
	find := func(c *mgo.Collection) error { return c.Find(query).Select(selector).All(result) }
	if err := m.Collection(collection, find); err != nil {
		return nil, err
	}
	return result, nil
}

// insert
func (m *mgoConfig) Insert(collection string, data interface{}) (err error) {
	find := func(c *mgo.Collection) error { return c.Insert(data) }
	if err = m.Collection(collection, find); err != nil {
		return
	}
	return
}

// delete
func (m *mgoConfig) Remove(collection string, query, selector, result interface{}) (interface{}, error) {
	find := func(c *mgo.Collection) error { return c.Find(query).Select(selector).All(result) }
	if err := m.Collection(collection, find); err != nil {
		return nil, err
	}
	return result, nil
}
