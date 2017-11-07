package model

import (
	"strconv"
	"net"
	"bufio"
	"fmt"

	"math"
	"io"
)

const (
	//ActionType 类型
	//ACTION_NULL = 0x03
	ACTION_PING byte = 0x01
	ACTION_SYNC_DATA byte = 0x02 //同步数据
	ACTION_CHUNK_DATA byte = 0x03 //同步数据，块数据
	ACTION_CHUNK_END byte = 0x04 //同步数据完成的标识


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

func GetDecodedPackageData(reader *bufio.Reader, conn net.Conn) *BackupPackage {
	//reader := bufio.NewReader(conn)

	packageData := new(BackupPackage)

	actionType,err := reader.ReadByte()
	checkErr(err)

	//logger.AsyncInfo(fmt.Sprintf("action: %#v", actionType))

	packageData.ActionType = actionType

	tempBuffer := make([]byte, DATA_LEGTH_TAG)
	n, err := reader.Read(tempBuffer)
	packageData.DataLength = bytesToInt32(tempBuffer)

	//err = binary.Read(reader, binary.LittleEndian, &packageData.DataLength)
	if err != nil {
		checkErr(fmt.Sprintf("获取包长度异常: err:%#v", err))
	}

	//logger.AsyncInfo(fmt.Sprintf("length:%#v", packageData.DataLength))

	//packageData.Data = make([]byte, packageData.DataLength)
	leftSize := packageData.DataLength

	for {
		buffer := make([]byte, int(math.Min(1024, float64(leftSize))))

		n, err = reader.Read(buffer)
		if err != nil && err != io.EOF {
			checkErr(fmt.Sprintf("获取数据异常:n:%#v , packge length:%d, err:%#v", n, packageData.DataLength, err))
		}

		packageData.Data = append(packageData.Data, buffer[0:n]...)

		leftSize = leftSize - int32(n)
		if leftSize <= 0 {
			break
		}
	}
	//n, err = reader.Read(packageData.Data)


	//logger.AsyncInfo(fmt.Sprintf("包数据: %#v, %#v, length:%d, %#v", actionType, 11111,packageData.DataLength, packageData.Data))

	return packageData
}

func NewBackupPackage(actionType byte) *BackupPackage {

	backupPackage := new(BackupPackage)

	backupPackage.ActionType = actionType

	return backupPackage
}

