package model

import "strconv"

const (
	//ActionType 类型
	ACTION_PING = 0x01
	ACTION_SYNC_DATA = 0x02 //同步数据

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

func NewBackupPackage(actionType byte) *BackupPackage {

	backupPackage := new(BackupPackage)

	backupPackage.ActionType = actionType

	return backupPackage
}

