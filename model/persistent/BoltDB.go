package persistent

import (
	//"strconv"
	"os"
	"github.com/boltdb/bolt"
)

var boltDb *bolt.DB

func GetBoltDB(dbFile string, mode os.FileMode, options *bolt.Options) *bolt.DB {

	//单例
	if boltDb != nil {
		return boltDb
	}

	var err error

	boltDb, err = bolt.Open(dbFile, mode, options)

	if err != nil {
		panic(err.Error())
	}

	return boltDb
}
