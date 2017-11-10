package main

//import  idGenerator "idGenerator/model"

import (
	"fmt"
	"flag"
	//"time"
	//"os"
	//"idGenerator/model/config"
	//"idGenerator/model/persistent"
	"github.com/gin-gonic/gin"
	"idGenerator/controller"
	"idGenerator/model"
	"idGenerator/model/logger"
)

//每个业务对应一个 key 全局唯一
//var idWorkerMap = make(map[int]*idGenerator.IdWorker)
//var idWorkerMap = cmap.New();

func main() {

	//初始化application
	application := model.GetApplication()

	//加载配置
	application.InitConfig("")
	configLog := fmt.Sprintf("loaded config %#v\napplication base_path:%s", application.ConfigData, application.BasePath)
	logger.AsyncInfo(configLog)

	port := "8182"

	//启动数据备份server
	flag.Parse()
	serverInstancType := flag.Arg(0)
	switch serverInstancType {
		case model.SERVER_MASTER:
			logger.AsyncInfo("启动备份server端程序")
			application.StartDataBackUpServer()

		case model.SERVER_SLAVE:

			port = "8183"
			application.ConfigData.ServerType = model.SERVER_SLAVE

			logger.AsyncInfo("启动slave端数据备份程序")
			application.ConfigData.Bolt.FilePath +=  ".backup"
			application.StartDataBackUpClient()

			logger.AsyncInfo("启动slave server数据通道")

		default:
			logger.AsyncInfo("输入参数:" + serverInstancType)
			panic("服务实例类型只能是master 或 slave")
	}

	//异步写log
	logger.AsyncInfo("application inited......")

	//r := gin.Default()
	r := gin.New()
	r.Use(logger.LoggerHanderFunc())
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Snow Flake算法
	r.GET("/snowflake/:id", controller.SnowFlakeAction)

	//自增方式
	r.GET("/autoincrement", controller.AutoIncrementAction)

	// Listen and Server in 0.0.0.0:8182
	r.Run(":" + port)

	//r.Run() // listen and serve on 0.0.0.0:8080
}
