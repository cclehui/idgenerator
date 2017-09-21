package model

import (
	"idGenerator/model/cmap"
)

var incrementIdWorkerInstance *IncrementIdWorker

//单例获取 递增方式的 id worker
func GetIncrementIdWorker() *IncrementIdWorker {
	if incrementIdWorkerInstance == nil {
		incrementIdWorkerInstance = new(IncrementIdWorker)
		incrementIdWorkerInstance.WorkerMap = cmap.New()
		incrementIdWorkerInstance.PersistType = GetApplication().ConfigData.PersistType
	}

	return incrementIdWorkerInstance
}
