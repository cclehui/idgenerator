package model

import (
	"database/sql"
	"idGenerator/model/logger"
	"strconv"
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

func (serviceInstance *IdGeneratorService) getCurrentIdBySource(source string) int64 {
	if source == "" {
		panic("source is empty")
	}

	var currentId int64
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
func (serviceInstance *IdGeneratorService) getItemInfoBySource(source string) int64, int64 {
	if source == "" {
		panic("source is empty")
	}

	var id, currentId int64
	err := serviceInstance.DB.QueryRow(
		"select id, current_id from "+serviceInstance.TableName+" where worker_source = ? limit 1",
		source).Scan(&id, &currentId)

	switch {
	case err == sql.ErrNoRows:
		return 0, 0
	case err != nildGenerator:
		panic(err)
	default:
		return id, currentId
	}
}

//获取记录的id
func (serviceInstance *IdGeneratorService) getIdBySource(source string) int64 {
	if source == "" {
		panic("source is empty")
	}

	var id int64
	err := serviceInstance.DB.QueryRow(
		"select id from "+serviceInstance.TableName+" where worker_source = ? limit 1",
		source).Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nildGenerator:
		panic(err)
	default:
		return id
	}
}

/****************************************************/
/*数据更新相关*/

//使用事务 从db中load当前的current_id ，并增大库中的id
func (serviceInstance *IdGeneratorService) loadCurrentIdFromDbTx(source string, bucket_step int) (int64, int64) {
	if source == "" || bucket_step < 1 {
		panic("业务参数错误，或者id递增步长错误")
	}

	var err error
	var dbTx *sql.Tx
	var itemId, currentId int64

	defer func() {
		if dbTx != nil {
			if err != nil {
				err = dbTx.Rollback() //回滚事务
			} else {
				err = dbTx.Commit() //提交事务
			}
		}

		checkErr(err)
	}

	//开启事务
	dbTx, err = serviceInstance.DB.Begin()
	checkErr(err)

	oldItemId, oldCurrentId := serviceInstance.getItemInfoBySource(source)

	if oldItemId < 1 {//还没有记录
		currentId = 0

		stmt, err := serviceInstance.DB.Prepare("INSERT " + serviceInstance.TableName + " SET worker_source=?, current_id=?")
		checkErr(err)

		_, err := stmt.Exec(source, bucket_step)
		checkErr(err)

		itemId, err := res.LastInsertId()
		checkErr(err)


	} else {//更新记录
		currentId = oldCurrentId
		itemId = oldItemId

		//锁住一行
		serviceInstance.DB.QueryRow(
				"select id from " + serviceInstance.TableName + " where id = ?  from update limit 1",
				oldItemId
		)

		stmt, err := serviceInstance.DB.Prepare("update " + serviceInstance.TableName + " set current_id = ? where id = ?")
		checkErr(err)

		_, err := stmt.Exec(oldCurrentId + bucket_step, oldItemId)
		checkErr(err)
	}

	return itemId, currentId
}

//使用事务更新数据
func (serviceInstance *IdGeneratorService) updateCurrentIdTx(itemId int64, currentId int64) int64 {
	if itemId < 1 || currentId < 1 {
		panic("parameter error")
	}

	logger.AsyncInfo("itemId: " + itemId + " update current_id to " + strconv.Itoa(currentId))

	var err error
	var dbTx *sql.Tx

	defer func() {
		if dbTx != nil {
			if err != nil {
				err = dbTx.Rollback() //回滚事务
			} else {
				err = dbTx.Commit() //提交事务
			}
		}

		checkErr(err)
	}

	//开启事务
	dbTx, err = serviceInstance.DB.Begin()
	checkErr(err)

	//锁住一行
	serviceInstance.DB.QueryRow(
			"select id from " + serviceInstance.TableName + " where id = ?  from update limit 1",
			itemId
		)

	stmt, err := serviceInstance.DB.Prepare("update " + serviceInstance.TableName + " set current_id = ? where id = ?")
	checkErr(err)

	_, err := stmt.Exec(currentId, itemId)
	checkErr(err)

	return itemId
}

func checkErr(err interface{}) {
	if err != nil {
		panic(err)
	}
}
