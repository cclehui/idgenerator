package model

import (
    "idGenerator/model/cmap"
    "errors"
)

type IncrementIdWorker struct {
    WorkerMap cmap.ConcurrentMap
}

type singleStorage struct {
    CurrentId int
    MaxId int
}

func (worker *IncrementIdWorker) NextId(source string) (int, error){
    if source == "" {
        return 0, errors.New("来源错误")
    }

    cachedStorage, hasOld := worker.WorkerMap.Get(source)

    var storage *singleStorage;

    if hasOld {
        tempStorage, typeOk := cachedStorage.(*singleStorage)
        if !typeOk {
            return 0, errors.New("旧数据类型异常")
        }

        storage = tempStorage

    } else {
        storage = &singleStorage{0, 10000}
        worker.WorkerMap.Set(source, storage)
    }

    storage.CurrentId = storage.CurrentId + 1

    return storage.CurrentId, nil
}




