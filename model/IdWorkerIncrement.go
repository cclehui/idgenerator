package model

import (
    "idGenerator/model/cmap"
    "errors"
)

//const (
    //BUCKET_STEP = 10000 //每次从db中拿到的递增量
//)

type IncrementIdWorker struct {
    WorkerMap cmap.ConcurrentMap
}

type singleStorage struct {
    CurrentId int
    CurrentMaxId int
}

//获取自增的唯一id db没有事务保证
func (worker *IncrementIdWorker) NextId(source string) (int, error){
    if source == "" {
        return 0, errors.New("来源错误")
    }

    cachedStorage, hasOld := worker.WorkerMap.Get(source)

    var storage *singleStorage;

    idGeneratorService := NewIdGeneratorService()

    if hasOld {
        tempStorage, typeOk := cachedStorage.(*singleStorage)
        if !typeOk {
            return 0, errors.New("旧数据类型异常")
        }

        storage = tempStorage

    } else {
        //从db中load 
        currentId := idGeneratorService.getCurrentIdBySource(source)
        currentMaxId := currentId + GetApplication().ConfigData.BucketStep

        // 当前最大的id持久化到db中
        idGeneratorService.updateSourceCurrentId(source, currentMaxId)

        storage = &singleStorage{currentId, currentMaxId}
        worker.WorkerMap.Set(source, storage)

    }

    storage.CurrentId = storage.CurrentId + 1

    if storage.CurrentId >= storage.CurrentMaxId {
        storage.CurrentMaxId += GetApplication().ConfigData.BucketStep

        // 当前最大的id增大 ， 并持久化到db中
        idGeneratorService.updateSourceCurrentId(source, storage.CurrentMaxId)
    }

    return storage.CurrentId, nil
}

//使用事务来持久化 
func (worker *IncrementIdWorker) NextIdWidthTx(source string) (int, error) {
    if source == "" {
        return 0, errors.New("来源错误")
    }

    cachedStorage, hasOld := worker.WorkerMap.Get(source)

    var storage *singleStorage;

    idGeneratorService := NewIdGeneratorService()

    if hasOld {
        tempStorage, typeOk := cachedStorage.(*singleStorage)
        if !typeOk {
            return 0, errors.New("旧数据类型异常")
        }

        storage = tempStorage

    } else {
        //从db中load 
        //cclehui_todo
        currentId := idGeneratorService.getCurrentIdBySource(source)
        currentMaxId := currentId + GetApplication().ConfigData.BucketStep

        // 当前最大的id持久化到db中
        idGeneratorService.updateSourceCurrentId(source, currentMaxId)

        storage = &singleStorage{currentId, currentMaxId}
        worker.WorkerMap.Set(source, storage)

    }

    storage.CurrentId = storage.CurrentId + 1

    if storage.CurrentId >= storage.CurrentMaxId {
        storage.CurrentMaxId += GetApplication().ConfigData.BucketStep

        // 当前最大的id增大 ， 并持久化到db中
        //cclehui_todo
        idGeneratorService.updateSourceCurrentId(source, storage.CurrentMaxId)
    }

    return storage.CurrentId, nil
}



