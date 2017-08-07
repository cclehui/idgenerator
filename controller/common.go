package controller

import idGenerator "idGenerator/model"

import (
    "github.com/gin-gonic/gin"
    "idGenerator/model/cmap"
    "strconv"
    "fmt"
)

func IdWorkerAction(idWorkerMap *cmap.ConcurrentMap, request *gin.Context ) {
    workerId := request.Params.ByName("id");
    currentWorker, ok := idWorkerMap.Get(workerId);
    value, typeOk := currentWorker.(idGenerator.IdWorker);

    if ok && typeOk {
        //获取下一个递增id
        nid, _ := value.NextId();

        request.JSON(200, gin.H{"id": nid})

    } else {

        id, _ := strconv.Atoi(workerId);

        idWorker, err := idGenerator.NewIdWorker(int64(id))
        if err == nil {
            nid, _ := idWorker.NextId();
            idWorkerMap.Set(workerId, idWorker);

            request.JSON(200, gin.H{"id": nid})

        } else {
            fmt.Println(err)
        }
    }
}
