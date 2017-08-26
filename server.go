package main

//import  idGenerator "idGenerator/model"

import(
    "fmt"
    //"time"
    //"os"
    //"idGenerator/model/config"
    //"idGenerator/model/persistent"
    "idGenerator/model/logger"
    "github.com/gin-gonic/gin"
    "idGenerator/model"
    "idGenerator/controller"
)


//每个业务对应一个 key 全局唯一
//var idWorkerMap = make(map[int]*idGenerator.IdWorker)
//var idWorkerMap = cmap.New();

func main() {

    //初始化application
    application := model.GetApplication();

    //加载配置
    application.InitConfig("");

    //异步写log
    logger.AsyncInfo("application inited......");

    configLog := fmt.Sprintf("loaded config %#v", application.ConfigData)

    logger.AsyncInfo(configLog);

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
    r.Run(":8182")

    //r.Run() // listen and serve on 0.0.0.0:8080
}
