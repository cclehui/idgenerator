package model

import (
	"idGenerator/model/cmap"
)

var autoincrIdWorkerInstance *AutoIncrIdWorker

//单例获取 递增方式的 id worker
func GetAutoIncrIdWorker() *AutoIncrIdWorker {
	if autoincrIdWorkerInstance == nil {
		autoincrIdWorkerInstance = new(AutoIncrIdWorker)
		autoincrIdWorkerInstance.WorkerMap = cmap.New()
		autoincrIdWorkerInstance.PersistType = GetApplication().ConfigData.PersistType
	}

	return autoincrIdWorkerInstance
}
