package model

import (
	"database/sql"
	"idGenerator/model/logger"
	"strconv"
	//"fmt"
)

type IdGeneratorService struct {
	DB        *sql.DB
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

	var currentId int
	err := serviceInstance.DB.QueryRow(
		"select current_id from "+serviceInstance.TableName+" where worker_source = ? limit 1",
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

//获取一条记录的信息
func (serviceInstance *IdGeneratorService) getItemInfoBySource(source string) (int, int) {
	if source == "" {
		panic("source is empty")
	}

	var id, currentId int
	err := serviceInstance.DB.QueryRow(
		"select id, current_id from "+serviceInstance.TableName+" where worker_source = ? limit 1",
		source).Scan(&id, &currentId)

	switch {
	case err == sql.ErrNoRows:
		return 0, 0
	case err != nil:
		panic(err)
	default:
		return id, currentId
	}
}

//获取记录的id
func (serviceInstance *IdGeneratorService) getIdBySource(source string) int {
	if source == "" {
		panic("source is empty")
	}

	var id int
	err := serviceInstance.DB.QueryRow(
		"select id from "+serviceInstance.TableName+" where worker_source = ? limit 1",
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

/****************************************************/
/*数据更新相关*/

//使用事务 从db中load当前的current_id ，并增大库中的id
func (serviceInstance *IdGeneratorService) loadCurrentIdFromDbTx(source string, bucket_step int) (int, int) {
	if source == "" || bucket_step < 1 {
		panic("业务参数错误，或者id递增步长错误")
	}

	logger.AsyncInfo("load current id from db, source: " + source + " , bucket_step: " + strconv.Itoa(bucket_step))

	var err error
	var dbTx *sql.Tx
	var itemId, currentId int

	defer func() {
		err := recover()

		if dbTx != nil {
			if err != nil {
				err = dbTx.Rollback() //回滚事务
			} else {
				err = dbTx.Commit() //提交事务
			}
		}

		checkErr(err)
	}()

	//开启事务
	dbTx, err = serviceInstance.DB.Begin()
	checkErr(err)

	oldItemId, oldCurrentId := serviceInstance.getItemInfoBySource(source)

	if oldItemId < 1 {//还没有记录
		currentId = 0

		stmt, err1 := serviceInstance.DB.Prepare("INSERT " + serviceInstance.TableName + " SET worker_source=?, current_id=?")
		defer stmt.Close()
		checkErr(err1)

		 res, err2 := stmt.Exec(source, bucket_step)
		 checkErr(err2)

		itemIdNew, err3 := res.LastInsertId()
		checkErr(err3)

		itemId = int(itemIdNew)

	} else {//更新记录
		currentId = oldCurrentId
		itemId = oldItemId

		//锁住一行
		serviceInstance.DB.QueryRow("select id from " + serviceInstance.TableName + " where id = ?  from update limit 1", oldItemId)

		stmt, err4 := serviceInstance.DB.Prepare("update " + serviceInstance.TableName + " set current_id = ? where id = ?")
		defer stmt.Close()
		checkErr(err4)

		_, err5 := stmt.Exec(int(oldCurrentId + bucket_step), oldItemId)
		checkErr(err5)
	}

	return itemId, currentId
}

//使用事务更新数据
func (serviceInstance *IdGeneratorService) updateCurrentIdTx(itemId int, currentId int, bucketStep int) (resultCurrentId int, newDbCurrentId int){
	if itemId < 1 || currentId < 1 {
		panic("parameter error")
	}

	logger.AsyncInfo("itemId: " + strconv.Itoa(itemId) + " update current_id to " + strconv.Itoa(currentId + bucketStep))

	var dbTx *sql.Tx
	var err error


	defer func() {
		err := recover()

		if dbTx != nil {
			if err != nil {
				err = dbTx.Rollback() //回滚事务
			} else {
				err = dbTx.Commit() //提交事务
			}
		}

		//异常没抛出来 cclehui_todo
		checkErr(err)
	}()

	//开启事务
	dbTx, err = serviceInstance.DB.Begin()
	checkErr(err)

	//锁住一行
	var dbCurrentId int

	err1 := serviceInstance.DB.QueryRow(
			"select current_id from " + serviceInstance.TableName + " where id = ? limit 1 for update ", 
			itemId).Scan(&dbCurrentId)

	checkErr(err1)

	newDbCurrentId = currentId + bucketStep
	resultCurrentId = currentId

	if dbCurrentId > currentId {
		resultCurrentId = dbCurrentId + 1
		newDbCurrentId = dbCurrentId + bucketStep;
	}

	stmt, err2 := serviceInstance.DB.Prepare("update " + serviceInstance.TableName + " set current_id = ? where id = ?")
	defer stmt.Close()
	checkErr(err2)

	_, err3 := stmt.Exec(newDbCurrentId, itemId)
	checkErr(err3)

	return resultCurrentId, newDbCurrentId
}

func checkErr(err interface{}) {
	if err != nil {
		panic(err)
	}
}
