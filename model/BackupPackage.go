package model

import (
	"strconv"
	"net"
	"bufio"
	"fmt"
)

const (
	//ActionType 类型
	//ACTION_NULL = 0x03
	ACTION_PING = 0x01
	ACTION_SYNC_DATA = 0x02 //同步数据

	DATA_LEGTH_TAG = 4

	MAX_DATA_LENGTH = 1000000 //最大的数据包长度

)

type BackupPackage struct {
	ActionType byte
	DataLength int32
	Data []byte
}

func (backupPackage *BackupPackage) encodeData(data []byte) {

	if len(data) > MAX_DATA_LENGTH {
		panic("数据包超长, 超过 " + strconv.Itoa(MAX_DATA_LENGTH))
	}

	backupPackage.Data = append(backupPackage.Data, data...)
	backupPackage.DataLength = int32(len(backupPackage.Data))

	return
}

func (backupPackage *BackupPackage) getHeader() []byte {
	result := make([]byte, 5);

	result = append(result, backupPackage.ActionType)
	result = append(result, int32ToBytes(int(backupPackage.DataLength))...)

	return result
}

func getPackageData(conn net.Conn) *BackupPackage {

	reader := bufio.NewReader(conn)

	packageData := new(BackupPackage)

	actionType,err := reader.ReadByte()
	checkErr(err)

	packageData.ActionType = actionType

	dataLength := make([]byte, DATA_LEGTH_TAG)
	n, err := reader.Read(dataLength)
	if n < DATA_LEGTH_TAG || err != nil {
		checkErr(fmt.Sprintf("解包长度异常:n:%#v, err:%#v", n, err))
	}

	//cclehui_todo




}

func NewBackupPackage(actionType byte) *BackupPackage {

	backupPackage := new(BackupPackage)

	backupPackage.ActionType = actionType

	return backupPackage
}

