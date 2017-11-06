package model

import (
	"strconv"
	"net"
	"bufio"
	"fmt"

	"encoding/binary"
)

const (
	//ActionType 类型
	//ACTION_NULL = 0x03
	ACTION_PING byte = 0x01
	ACTION_SYNC_DATA byte = 0x02 //同步数据
	ACTION_CHUNK_DATA byte = 0x03 //同步数据，块数据

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

	result[0] = backupPackage.ActionType
	for index,value := range int32ToBytes(int(backupPackage.DataLength)) {
		result[index + 1] = value
	}
	return result
}

func GetDecodedPackageData(conn net.Conn) *BackupPackage {
	reader := bufio.NewReader(conn)

	packageData := new(BackupPackage)

	actionType,err := reader.ReadByte()
	checkErr(err)

	//logger.AsyncInfo(fmt.Sprintf("action: %#v", actionType))

	packageData.ActionType = actionType

	//tempBuffer := make([]byte, DATA_LEGTH_TAG)
	//n, err := reader.Read(tempBuffer)
	//packageData.DataLength = int32(bytesToInt(tempBuffer))

	err = binary.Read(reader, binary.BigEndian, &packageData.DataLength)
	if err != nil {
		checkErr(fmt.Sprintf("获取包长度异常: err:%#v", err))
	}

	//logger.AsyncInfo(fmt.Sprintf("length:%#v", packageData.DataLength))

	packageData.Data = make([]byte, packageData.DataLength)

	n, err := reader.Read(packageData.Data)

	if n < int(packageData.DataLength) || err != nil {
		checkErr(fmt.Sprintf("获取数据异常:n:%#v, err:%#v", n, err))
	}

	return packageData
}

func NewBackupPackage(actionType byte) *BackupPackage {

	backupPackage := new(BackupPackage)

	backupPackage.ActionType = actionType

	return backupPackage
}

