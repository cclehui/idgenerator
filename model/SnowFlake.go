// This package provides unique id in distribute system
// the algorithm is inspired by Twitter's famous snowflake
// its link is: https://github.com/twitter/snowflake/releases/tag/snowflake-2010
//

// +---------------+----------------+----------------+
// |timestamp(ms)42  | worker id(10) | sequence(12)	 |
// +---------------+----------------+----------------+

// Copyright (C) 2017 by cclehui

package model

import (
	"errors"
	"sync"
	"time"
	//"fmt"
    "idGenerator/model/logger"
)

const (
	CEpoch         = 1474802888000
	CWorkerIdBits  = 10 // Num of WorkerId Bits
	CSenquenceBits = 12 // Num of Sequence Bits

	CWorkerIdShift  = 12
	CTimeStampShift = 22

	CSequenceMask = 0xfff // equal as getSequenceMask()
	CMaxWorker    = 0x3ff // equal as getMaxWorkerId()
)

// SnowFlakeIdWorker Struct
type SnowFlakeIdWorker struct {
	workerId      int64
	lastTimeStamp int64
	sequence      int64
	maxWorkerId   int64
	lock          *sync.Mutex
}

// NewSnowFlakeIdWorker Func: Generate NewSnowFlakeIdWorker with Given workerid
func NewSnowFlakeIdWorker(workerid int64) (iw *SnowFlakeIdWorker, err error) {
	maxWorkerId := getMaxWorkerId()

	if workerid > maxWorkerId || workerid < 0 {
		return nil, errors.New("workerid 异常")
	}

    logger.Printf("NewSnowFlakeIdWorker, max worker ID:%#v, current:%#v\n", maxWorkerId, workerid);

	iw = new(SnowFlakeIdWorker)
    iw.maxWorkerId = maxWorkerId
	iw.workerId = workerid
	iw.lastTimeStamp = -1
	iw.sequence = 0
	iw.lock = new(sync.Mutex)

	return iw, nil
}

func getMaxWorkerId() int64 {
	return -1 ^ -1<<CWorkerIdBits
}

func getSequenceMask() int64 {
	return -1 ^ -1<<CSenquenceBits
}

// return in ms
func (iw *SnowFlakeIdWorker) timeGen() int64 {
	return time.Now().UnixNano() / 1000 / 1000
}

func (iw *SnowFlakeIdWorker) timeReGen(last int64) int64 {
	ts := time.Now().UnixNano()
	for {
		if ts < last {
			ts = iw.timeGen()
		} else {
			break
		}
	}
	return ts
}

// NewId Func: Generate next id
func (iw *SnowFlakeIdWorker) NextId() (ts int64, err error) {
	iw.lock.Lock()
	defer iw.lock.Unlock()
	ts = iw.timeGen()
	if ts == iw.lastTimeStamp {
		iw.sequence = (iw.sequence + 1) & CSequenceMask
		if iw.sequence == 0 {
			ts = iw.timeReGen(ts)
		}
	} else {
		iw.sequence = 0
	}

	if ts < iw.lastTimeStamp {
		err = errors.New("Clock moved backwards, Refuse gen id")
		return 0, err
	}
	iw.lastTimeStamp = ts
	ts = (ts-CEpoch)<<CTimeStampShift | iw.workerId<<CWorkerIdShift | iw.sequence
	return ts, nil
}

// ParseId Func: reverse uid to timestamp, workid, seq
func ParseId(id int64) (t time.Time, ts int64, workerId int64, seq int64) {
	seq = id & CSequenceMask
	workerId = (id >> CWorkerIdShift) & CMaxWorker
	ts = (id >> CTimeStampShift) + CEpoch
	t = time.Unix(ts/1000, (ts%1000)*1000000)
	return
}
