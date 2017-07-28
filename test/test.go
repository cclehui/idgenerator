package main

import(
    "fmt"
    //"time"
    //"os"
    "strconv"
    "idGenerator"
    "idGenerator/config"
    "idGenerator/persistent"
    "github.com/gin-gonic/gin"
    "github.com/cmap"
)

//每个业务对应一个 key 全局唯一
//var idWorkerMap = make(map[int]*idGenerator.IdWorker)
var idWorkerMap = cmap.New();

func main() {

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
    r.GET("/worker/:id", func(c *gin.Context) {
        workerId := c.Params.ByName("id");
        currentWorker, ok := idWorkerMap.Get(workerId);
        value, typeOk := currentWorker.(idGenerator.IdWorker);

        if ok && typeOk {
            //获取下一个递增id
            nid, _ := value.NextId();

            c.JSON(200, gin.H{"id": nid})

        } else {

            id, _ := strconv.Atoi(workerId);

            idWorker, err := idGenerator.NewIdWorker(int64(id))
            if err == nil {
                nid, _ := idWorker.NextId();
                idWorkerMap.Set(workerId, idWorker);

                c.JSON(200, gin.H{"id": nid})

            } else {
                fmt.Println(err)
            }
        }
    })

    // Listen and Server in 0.0.0.0:8182
    //r.Run(":8182")

    r.Run() // listen and serve on 0.0.0.0:8080
}
