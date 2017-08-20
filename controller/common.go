package controller

//import idGenerator "idGenerator/model"

import (
    "github.com/gin-gonic/gin"
    //"idGenerator/model/cmap"
    "idGenerator/model"
    //"idGenerator/model/logger"
    "idGenerator/model/jsonApi"
    "strconv"
    //"fmt"
)

func IdWorkerAction(context *gin.Context ) {
    workerSource := context.Params.ByName("id");
    workerid, err := strconv.Atoi(workerSource);

    if err != nil {
        jsonApi.Fail(context, err.Error(), 100001)
        return
    }

    idWorkerMap := model.GetApplication().GetIdWorkerMap();
    currentWorker, hasOld := idWorkerMap.Get(workerSource);

    if hasOld {
        //获取下一个递增id
        workerInstance, typeOk := currentWorker.(*model.SnowFlakeIdWorker);
        if !typeOk {
            jsonApi.Fail(context, "workerInstance类型错误", 100002)
            return;
        }

        nid, _ := workerInstance.NextId();
        jsonApi.Success(context, gin.H{"id": nid})
        return
    }

    //获取新的
    workerInstance, err := model.NewSnowFlakeIdWorker(int64(workerid))
    if err != nil {
        jsonApi.Fail(context, err.Error(), 100001)
        return
    }

    nid, _ := workerInstance.NextId();
    idWorkerMap.Set(workerSource, workerInstance);
    jsonApi.Success(context, gin.H{"id": nid})
}
