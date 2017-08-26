package model

import (
    "database/sql"
)

type IdGeneratorService struct {
    DB *sql.DB
    TableName string
}

func NewIdGeneratorService() *IdGeneratorService {
    var serviceInstance = new(IdGeneratorService)

    db, err := GetApplication().GetMysqlDB()
    checkErr(err)

    serviceInstance.DB = db
    serviceInstance.TableName = "idGenerator"

    return serviceInstance
}

func (serviceInstance *IdGeneratorService) getCurrentIdBySource(source string) int {
    if source == "" {
        panic("source is empty")
    }

    var currentId int;
    err := serviceInstance.DB.QueryRow(
                "select current_id from " + serviceInstance.TableName + " where worker_source = ? limit 1", 
                    source).Scan(&currentId)

    switch {
         case err == sql.ErrNoRows:
              return 0
         case err != nil:
              panic(err)
         default:
            return currentId
    }
}

func (serviceInstance *IdGeneratorService) getIdBySource(source string) int {
    if source == "" {
        panic("source is empty")
    }

    var id int;
    err := serviceInstance.DB.QueryRow(
                "select id from " + serviceInstance.TableName + " where worker_source = ? limit 1", 
                    source).Scan(&id)

    switch {
         case err == sql.ErrNoRows:
              return 0
         case err != nil:
              panic(err)
         default:
            return id
    }
}

//更新数据
func (serviceInstance *IdGeneratorService) updateSourceCurrentId(source string, currentId int) int {
    if source == "" || currentId < 1 {
        panic("parameter error")
    }

    oldId := serviceInstance.getIdBySource(source)

    if oldId > 0 { //更新数据
        stmt, err := serviceInstance.DB.Prepare("update " + serviceInstance.TableName +" set current_id = ? where worker_source = ?")
        checkErr(err)

        _, err2 := stmt.Exec(currentId, source)
        checkErr(err2)

        return oldId

    } else {
        stmt, err := serviceInstance.DB.Prepare("INSERT " + serviceInstance.TableName + " SET worker_source=?,current_id=?") 
        checkErr(err)

        res, err := stmt.Exec(source, currentId)
        checkErr(err)

        id, err := res.LastInsertId() 
        checkErr(err)

        return int(id)
    }
}

func checkErr(err interface{}) {
    if err != nil {
        panic(err)
    }
}

