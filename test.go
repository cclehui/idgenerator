package main

//import  idGenerator "idGenerator/model"

import(
    "fmt"
    //"time"
    //"os"
    //"idGenerator/model/config"
    //"idGenerator/model/persistent"
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

    fmt.Printf("application inited:%#v\n", application.ConfigData);

    r := gin.Default()

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    // Get ID
    r.GET("/worker/:id", controller.IdWorkerAction)

    // Listen and Server in 0.0.0.0:8182
    r.Run(":8182")

    //r.Run() // listen and serve on 0.0.0.0:8080
}
