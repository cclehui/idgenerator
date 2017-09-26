package config

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

type Config struct {
	Addr           string `toml: "addr"`
	LogPath        string `toml: "log_path"`
	LogLevel       string `toml: "logLevel"`
	PersistType    int    `toml: "persistType"`
	DataDir        string    `toml: "dataDir"`
	BucketStep     int    `toml: "bucketStep"`
	ServerType     int    `toml: "serverType"`
	MasterAddress  string      `toml: "masterAddress"`
	UseTransAction bool   `toml: "useTransAction"`
	Bolt           Bolt  `toml: "bolt"`
	Mysql          Mysql  `toml: "mysql"`
}

type Bolt struct {
	FilePath         string `toml: "filePath"`
	BucketName       string `toml: "bucketName"`
}

type Mysql struct {
	Host         string `toml: "host"`
	Port         int    `toml: "port"`
	Name         string `toml: "name"`
	User         string `toml: "user"`
	Password     string `toml: "password"`
	MaxIdleConns int    `toml: "maxIdleConns"`
	MaxOpenConns int    `toml: "maxOpenConns"`
}

func GetConfigFromFile(configFile string) Config {
	if configFile == "" {
		panic("配置文件不存在")
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic("配置文件不存在")
	}

	//配置文件
	var config Config

	if _, err := toml.Decode(string(data), &config); err != nil {
		panic("配置格式错误")
	}

	return config
}
