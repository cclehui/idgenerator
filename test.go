package main

//import  idGenerator "idGenerator/model"

import(
    "fmt"
    //"time"
    //"os"
    "idGenerator/model/config"
    "idGenerator/model/persistent"
    "github.com/gin-gonic/gin"
    "idGenerator/model/cmap"
    "idGenerator/controller"
)


//每个业务对应一个 key 全局唯一
//var idWorkerMap = make(map[int]*idGenerator.IdWorker)
var idWorkerMap = cmap.New();

func main() {
    fmt.Println("11111111111");

    config := config.GetInstance("");

    db := persistent.GetMysqlDB(
                config.Mysql.User,
                config.Mysql.Password,
                config.Mysql.Host,
                config.Mysql.Port,
                config.Mysql.Name,
            )

    fmt.Printf("%#v\n", config.Addr);
    fmt.Printf("%#v\n", db);
    return;

    r := gin.Default()

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    // Get ID
    r.GET("/worker/:id", controller.IdWorkerAction)

    // Listen and Server in 0.0.0.0:8182
    //r.Run(":8182")

    r.Run() // listen and serve on 0.0.0.0:8080
}
