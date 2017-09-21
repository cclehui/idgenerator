package model

import (
	"errors"
	//"strconv"
	"idGenerator/model/cmap"
	"idGenerator/model/logger"
)

const (
	//BUCKET_STEP = 10000 //每次从db中拿到的递增量
	PERSIST_TYPE_MYSQL = 1 //mysql持久化
	PERSIST_TYPE_BOLTDB = 2 //boltdb持久化
)

type IncrementIdWorker struct {
	WorkerMap cmap.ConcurrentMap
	PersistType int
}

type singleStorage struct {
	ItemId    int
	CurrentId    int
	CurrentMaxId int
}

//获取递增id
func (worker *IncrementIdWorker) NextId(source string) (result int, err error) {
	if worker.PersistType == PERSIST_TYPE_BOLTDB {
		//Boltdb 持久化
		result, err = worker.NextIdByBoltDb(source)

	} else {

		//mysql 持久化
		result, err = worker.NextIdWidthTx(source)
	}
	return
}

//使用boltdb持久化
func (worker *IncrementIdWorker) NextIdByBoltDb(source string) (int, error) {
	if source == "" {
		return 0, errors.New("来源错误")
	}

	//cachedStorage, hasOld := worker.WorkerMap.Get(source)

	result := NewBoltDbIdGenerator().NextId(source)


	return result, nil
}

//使用mysql事务来持久化
func (worker *IncrementIdWorker) NextIdWidthTx(source string) (int, error) {
	if source == "" {
		return 0, errors.New("来源错误")
	}

	cachedStorage, hasOld := worker.WorkerMap.Get(source)

	var storage *singleStorage

	idGeneratorService := NewIdGeneratorService()

	if hasOld {//内存中有
		tempStorage, typeOk := cachedStorage.(*singleStorage)
		if !typeOk {
			return 0, errors.New("旧数据类型异常")
		}

		storage = tempStorage

	} else {
		//从db中load
		itemId, currentId := idGeneratorService.loadCurrentIdFromDbTx(source, GetApplication().ConfigData.BucketStep)
		currentMaxId := currentId + GetApplication().ConfigData.BucketStep


		storage = &singleStorage{itemId, currentId, currentMaxId}
		worker.WorkerMap.Set(source, storage)

	}

	storage.CurrentId = storage.CurrentId + 1

	//当前id超过内存中允许的最大值了 需要增大最大值， 并持久化到db中
	if storage.CurrentId >= storage.CurrentMaxId {

		newCurrentId, newMaxId := idGeneratorService.updateCurrentIdTx(storage.ItemId, storage.CurrentId, GetApplication().ConfigData.BucketStep)
		logger.AsyncInfo(newCurrentId)
		logger.AsyncInfo(newMaxId)

		storage.CurrentId = newCurrentId
		storage.CurrentMaxId = newMaxId
	}

	return storage.CurrentId, nil
}
