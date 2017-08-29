package model

import (
	"database/sql"
	"idGenerator/model/cmap"
	"idGenerator/model/config"
	"idGenerator/model/persistent"
)

type Application struct {
	IdWorkerMap cmap.ConcurrentMap // 应用的处理worker
	ConfigData  config.Config      //配置信息
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
func (application *Application) InitConfig(configPath string) {
	application.ConfigData = config.GetInstance(configPath)
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

//获取worker map
func (application *Application) GetIdWorkerMap() cmap.ConcurrentMap {
	applicationInstance := GetApplication()

	return applicationInstance.IdWorkerMap
}
