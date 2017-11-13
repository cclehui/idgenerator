package model

import (
	"errors"
	"fmt"
)

type BoltDbRpcService  struct {
	BoltDbService *BoltDbService
}

func NewBoltDbRpcService() *BoltDbRpcService {
	return &BoltDbRpcService{NewBoltDbService()}
}

type loadCurrentIdFromDbArgs struct {
	Source string
	BucketStep int
}

func (this *BoltDbRpcService) loadCurrentIdFromDb(args *loadCurrentIdFromDbArgs, result *int) (err error) {

	defer func() {
		errRecovered := recover()

		if errRecovered != nil {
			err = errors.New(fmt.Sprintf("%#v", errRecovered))
		}
	}()

	*result = this.BoltDbService.loadCurrentIdFromDb(args.Source, args.BucketStep)
	return err
}

type IncrSourceCurrentIdArgs struct {
	Source string
	CurrentId int
	BucketStep int
}

type IncrSourceCurrentIdResult struct {
	ResultCurrentId	int
	NewDbCurrentId int
}

func (this *BoltDbRpcService) IncrSourceCurrentId(args *IncrSourceCurrentIdArgs, result *IncrSourceCurrentIdResult) (err error) {

	defer func() {
		errRecovered := recover()

		if errRecovered != nil {
			err = errors.New(fmt.Sprintf("%#v", errRecovered))
		}
	}()

	resultCurrentId, newDbCurrentId := this.BoltDbService.IncrSourceCurrentId(args.Source, args.CurrentId, args.BucketStep)

	result = &IncrSourceCurrentIdResult{resultCurrentId, newDbCurrentId}

	return err
}