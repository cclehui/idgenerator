package model

import (
	"idGenerator/model/logger"
	"github.com/boltdb/bolt"
	"strconv"
	"bytes"
	"encoding/binary"
)

const (
	BUCKET_NAME = "IdGeneratorBucket"
)

type BoltDbService struct {
	BucketName string
}

func NewBoltDbService() *BoltDbService {

	boltDb, err := GetApplication().GetBoltDB()
	CheckErr(err)

	boltDb.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_NAME))
			CheckErr(err)
			return nil
	})

	return &BoltDbService{BUCKET_NAME}
}

func (this *BoltDbService) NextId(source string) int {
	logger.AsyncInfo(source)

	boltDb, err := GetApplication().GetBoltDB()
	CheckErr(err)

	boltDb.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(BUCKET_NAME))
			err := b.Put([]byte(source), []byte("100"))
			CheckErr(err)

			return nil
	})

	return 100
}

/****************************************************/
/*数据更新相关*/

//使用事务 从db中load当前的current_id ，并增大库中的id
func (this *BoltDbService) loadCurrentIdFromDb(source string, bucketStep int) int {
	if source == "" || bucketStep < 1 {
		panic("业务参数错误，或者id递增步长错误")
	}

	var currentId int

	boltDb, errGetBolt := GetApplication().GetBoltDB()
	CheckErr(errGetBolt)

	//开启事务
	dbTx, errTx := boltDb.Begin(true)
	checkErr(errTx)

	defer func() {
		err := recover()

		if dbTx != nil {
			if err != nil {
				dbTx.Rollback() //回滚事务
			} else {
				dbTx.Commit() //提交事务
			}
		}

		checkErr(err)
	}()

	bucket := dbTx.Bucket([]byte(this.BucketName))

	oldCurrentId := bucket.Get([]byte(source))

	if oldCurrentId == nil {//还没有记录

		currentId = 0

		errInsert := bucket.Put([]byte(source), intToBytes(currentId +bucketStep))
		checkErr(errInsert)

	} else {//更新记录

		currentId = bytesToInt(oldCurrentId)

		errUpdate := bucket.Put([]byte(source), intToBytes(currentId +bucketStep))
		checkErr(errUpdate)
	}

	logger.AsyncInfo("load current id from boltdb, source: " + source + " , currentId: " + strconv.Itoa(currentId))

	return currentId
}

//使用事务更新数据
func (this *BoltDbService) IncrSourceCurrentId(source string, currentId int, bucketStep int) (resultCurrentId int, newDbCurrentId int){
	if currentId < 1 || bucketStep < 1 {
		panic("parameter error")
	}

	boltDb, errGetBolt := GetApplication().GetBoltDB()
	CheckErr(errGetBolt)

	//开启事务
	dbTx, errTx := boltDb.Begin(true)
	checkErr(errTx)

	defer func() {
		err := recover()

		if dbTx != nil {
			if err != nil {
				dbTx.Rollback() //回滚事务
			} else {
				dbTx.Commit() //提交事务
			}
		}

		checkErr(err)
	}()

	bucket := dbTx.Bucket([]byte(this.BucketName))

	dbRes := bucket.Get([]byte(source))
	if dbRes == nil {//还没有记录
		panic("boltdb中数据不存在, 不可更新")
	}

	oldCurrentId := bytesToInt(dbRes)

	resultCurrentId = currentId
	newDbCurrentId = currentId + bucketStep

	if oldCurrentId > currentId {
		resultCurrentId = oldCurrentId + 1
		newDbCurrentId = oldCurrentId + bucketStep;
	}

	errUpdate := bucket.Put([]byte(source), intToBytes(newDbCurrentId))
	checkErr(errUpdate)

	logger.AsyncInfo("source: " + source + " update bolt current_id to " + strconv.Itoa(newDbCurrentId))

	return resultCurrentId, newDbCurrentId
}

func (this *BoltDbService) CallFuncFromMaster() {
	
}

//整形转换成字节  
func intToBytes(n int) []byte {
    bytesBuffer := bytes.NewBuffer([]byte{})
	tmp := int64(n)
    binary.Write(bytesBuffer, binary.LittleEndian, tmp)
    return bytesBuffer.Bytes()
}

//字节转换成整形  
func bytesToInt(b []byte) int {
    bytesBuffer := bytes.NewBuffer(b)
    var tmp int64
    binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
    return int(tmp)
}
