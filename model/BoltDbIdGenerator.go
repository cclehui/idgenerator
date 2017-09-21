package model

import (
	"idGenerator/model/logger"
	"github.com/boltdb/bolt"
)

const (
	BUCKET_NAME = "IdGeneratorBucket"
)

type BoltDbIdGenerator struct {
	BucketName string
}

func NewBoltDbIdGenerator() *BoltDbIdGenerator {
	
	boltDb, err := GetApplication().GetBoltDB()
	CheckErr(err)

	boltDb.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_NAME))
			CheckErr(err)
			return nil
	})

	return &BoltDbIdGenerator{BUCKET_NAME}
}

func (this *BoltDbIdGenerator) NextId(source string) int {
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
