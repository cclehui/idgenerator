package config

import (
    "io/ioutil"
    "github.com/BurntSushi/toml"
);

type Config struct {
    Addr string `toml: "addr"`
    LogPath string `toml: "log_path"`
    LogLevel string `toml: "log_level"`
    Mysql Mysql `toml: "mysql"`
}

type Mysql struct {
    Host string `toml: "host"`
    Port int `toml: "port"`
    Name string `toml: "name"`
    User string `toml: "user"`
    Password string `toml: "password"`
    MaxIdleConns int `toml: "max_idle_conns"`
    MaxOpenConns int `toml: "max_open_conns"`
}

func GetInstance(configFile string) Config {
    if configFile == "" {
        configFile = "./config/production.toml";
    }

    data, err := ioutil.ReadFile(configFile);
    if err != nil {
        panic("配置文件不存在")
    }

    //配置文件
    var config Config;

    if _, err := toml.Decode(string(data), &config); err != nil {
        panic("配置格式错误");
    }

    return config;
}
