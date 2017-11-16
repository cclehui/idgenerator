package model

import (
	"database/sql"
	"os"
	"fmt"
	"time"
	"idGenerator/model/cmap"
	"idGenerator/model/config"
	"idGenerator/model/persistent"
	"idGenerator/model/logger"
	"github.com/boltdb/bolt"
)

const (
	SERVER_MASTER = "master"
	SERVER_SLAVE = "slave"
)

type Application struct {
	IdWorkerMap cmap.ConcurrentMap // 应用的处理worker
	ConfigData  config.Config      //配置信息
	ConfigFileInfo os.FileInfo     //配置文件的文件信息
	BasePath string //应用根目录
	DataBackUpSocketClient *Client
	RpcSocketClient *Client
}

var application *Application
var applicationInited bool = false

//获取application 实例
func GetApplication() *Application {
	if !applicationInited {
		application = new(Application)
		application.IdWorkerMap = cmap.New()
		applicationInited = true
	}

	return application
}

//初始配置
func (application *Application) InitConfig(configFile string) {
	if configFile == "" {
		configFile = "./config/production.toml"
	}

	//应用根目录
	if pwd, err := os.Getwd(); err != nil {
		panic(err)

	} else {
		application.BasePath = pwd
	}

	//配置文件信息
	fileInfo, err := os.Stat(configFile)
	if err != nil {
		panic(err)
	}

	application.ConfigFileInfo = fileInfo

	application.ConfigData = config.GetConfigFromFile(configFile)

	//异步 如果配置文件有修改, 动态load 配置文件
	go func() {

		for {
			time.Sleep(1000 * time.Millisecond)

			fileInfoTemp, err := os.Stat(configFile)
			if err != nil {
				logger.AsyncInfo(fmt.Sprintf("配置文件热加载, stat err, %#v", err))
				continue
			}

			if fileInfoTemp.ModTime().Equal(application.ConfigFileInfo.ModTime()) {
				//文件没修改
				continue;
			}

			application.ConfigFileInfo = fileInfoTemp

			logger.AsyncInfo("配置文件热加载, start...")

			waitChan := make(chan bool)

			go func() {
				defer func() {
					err := recover()
					if err != nil {
						logger.AsyncInfo(fmt.Sprintf("配置文件热加载异常, %#v", err))
					}

					waitChan<-true
				}()

				application.ConfigData = config.GetConfigFromFile(configFile)

			}()

			<-waitChan

			logger.AsyncInfo("配置文件热加载, end...")
		}

	}()
}

//启动数据备份服务
func (application *Application) StartDataBackUpServer() {

	go func() {
		masterServer := NewServer(application.ConfigData.MasterAddress, SERVER_TYPE_DATA_BACKUP)
		masterServer.StartMasterServer()
	}()
}

//启动数据备份client
func (application *Application) StartDataBackUpClient() {
	defer func() {
		err := recover()

		logger.AsyncInfo(fmt.Sprintf("连接master 异常, %#v", err))
		if err != nil {
			panic(err)
		}

	}()

	//获取连接
	client := NewClient(application.ConfigData.MasterAddress)

	go func() {
		client.StartClientBackUp()
		application.DataBackUpSocketClient = client
	}()
}

//启动rpc server端
func (application *Application) StartRpcServer() {

	go func() {
		masterServer := NewServer(application.ConfigData.RpcSeverAddress, SERVER_TYPE_RPC)
		masterServer.StartMasterServer()
	}()

}

//启动 rpc client
func (application *Application) StartRpcClient() {
	defer func() {
		err := recover()

		logger.AsyncInfo(fmt.Sprintf("连接rpc server 异常, %#v", err))
		if err != nil {
			panic(err)
		}

	}()

	//获取连接
	client := NewClient(application.ConfigData.RpcSeverAddress)

	go func() {
		client.StartRpcClient()
		application.RpcSocketClient = client
	}()
}

//获取Mysql连接
func (application *Application) GetMysqlDB() (db *sql.DB, err interface{}) {
	defer func() {
		err = recover()
		return
	}()

	db = persistent.GetMysqlDB(
		application.ConfigData.Mysql.User,
		application.ConfigData.Mysql.Password,
		application.ConfigData.Mysql.Host,
		application.ConfigData.Mysql.Port,
		application.ConfigData.Mysql.Name,
		application.ConfigData.Mysql.MaxIdleConns,
		application.ConfigData.Mysql.MaxOpenConns,
	)

	return db, nil
}

//获取BoltDB
func (application *Application) GetBoltDB() (db *bolt.DB, err interface{}) {
	defer func() {
		err = recover()
		return
	}()

	db = persistent.GetBoltDB(application.ConfigData.Bolt.FilePath, 0777, &bolt.Options{Timeout: 30 * time.Second})

	return db, nil
}

//获取worker map
func (application *Application) GetIdWorkerMap() cmap.ConcurrentMap {
	applicationInstance := GetApplication()

	return applicationInstance.IdWorkerMap
}



