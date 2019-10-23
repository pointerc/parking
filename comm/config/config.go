package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct{
	Name string
}

func Init(s string) error {
	c := Config{
		Name: s,
	}
	if err := c.LoadConfigFile(); err != nil {
		return err
	}
	return nil
}

func (c *Config) LoadConfigFile() error {
	if c.Name != "" {
		//如果指定了配置文件，则解析配置文件
		viper.SetConfigFile(c.Name)
	} else {
		viper.AddConfigPath("$GOPATH/src/parking/configs")
		viper.SetConfigName("system")
	}
	viper.SetConfigType("yaml")
	//viper解析配置文件
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("解析配置文件失败：", err)
		return err
	}
	return nil

}
