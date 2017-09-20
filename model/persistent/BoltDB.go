package persistent

import (
	"strconv"
	"os"
	"github.com/boltdb/bolt"
)

var db *bolt.DB

func GetBoltDB(dbFile string, mode os.FileMode, options *bolt.Options) *bolt.DB {

	//单例
	if db != nil {
		return db
	}

	var err error

	db, err = bolt.Open(dbFile, mode, options)

	if err != nil {
		panic(err.Error())
	}

	return db
}
