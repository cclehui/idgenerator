package model;

import (
    "idGenerator/model/cmap"
);

type Application struct {
    idWorkerMap cmap.ConcurrentMap
    inited bool
}

var application *Application;

//获取application 实例
func GetApplication() *Application {
    if !application.inited  {
        application = new(Application);
    }

    return application;
}

//获取worker map
func (application *Application) GetIdWorkerMap() cmap.ConcurrentMap {
    applicationInstance := GetApplication();

    return applicationInstance.idWorkerMap;
}
