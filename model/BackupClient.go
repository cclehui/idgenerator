package model

import (
	"net"
	"time"
	"idGenerator/model/logger"
	"fmt"
	"sync"
	"os"
)

//var	contextList *list.List

type Client struct {
	Context *Context
	MasterAddress string
}

var client *Client

//启动client 备份
func StartClientBackUp(masterAddress string) {
	tcpAddress, err := net.ResolveTCPAddr("tcp", masterAddress)
	CheckErr(err)

	//connection, err := net.Dial("tcp", masterAddress)
	connection, err := net.DialTCP("tcp", nil, tcpAddress)
	CheckErr(err)

	now := time.Now().Unix()
	lock := new(sync.Mutex)
	var context = &Context{connection, now,lock,nil,nil}

	if client == nil {
		client = &Client{context, masterAddress}
	}

	//备份数据库
	go func() {
		for {
			logger.AsyncInfo("启动主从同步操作")
			client.doAction()

			time.Sleep(2 * time.Second)
		}
	}()
}

func (client *Client) doAction() {
	defer func() {
		err := recover()
		logger.AsyncInfo("重连master 异常捕获")
		logger.AsyncInfo(err)
	}()

	defer func() {
		err := recover()
		if err != nil {
			logger.AsyncInfo(err)

			if _, ok := err.(*net.OpError); ok {
				logger.AsyncInfo("重连master")
				client.Context.Connection.Close()
				client.reConnect()//尝试重连
			}
		}
	}()

	go client.sendHeartBeat()

	syncDataMsgChan := make(chan bool)
	//发送数据备份的请求
	go client.sendSyncDatabaseRequest(syncDataMsgChan)

	syncDataMsgChan <- true

	//读数据
	var backupDataFile *os.File = nil
	var err error
	var totalSize int64 = 0
	count := 0
	for {
		count++

		dataPackage := GetDecodedPackageData(client.Context.getReader(), client.Context.Connection)
		logger.AsyncInfo(fmt.Sprintf("开始解包: count:%d, action: %#v, length:%d", count, dataPackage.ActionType, dataPackage.DataLength))

		switch dataPackage.ActionType {
		case ACTION_PING:
			logger.AsyncInfo("心跳包返回")
		case ACTION_SYNC_DATA:
			// 重复写入一个文件 xxxxxxxxxxxxxx  cclehui_todo

			backupDataFile, err = os.OpenFile(GetApplication().ConfigData.Bolt.FilePath, os.O_WRONLY|os.O_CREATE, 0644)
			checkErr(err)

			err = backupDataFile.Truncate(0)
			checkErr(err)

			n, err := backupDataFile.Write(dataPackage.Data)
			checkErr(err)
			//backupDataFile.Close()

			//重新以append 方式打开文件
			//backupDataFile, err = os.OpenFile(GetApplication().ConfigData.Bolt.FilePath, os.O_WRONLY|os.O_APPEND, 0644)
			//defer backupDataFile.Close()
			//checkErr(err)

			totalSize = int64(n)

		case ACTION_CHUNK_DATA:
			if backupDataFile != nil {
				n, err := backupDataFile.Write(dataPackage.Data)
				checkErr(err)

				totalSize += int64(n)
			}
		case ACTION_CHUNK_END:
			logger.AsyncInfo(fmt.Sprintf("同步完成， 共同步数据 : %d bytes", totalSize ))
			backupDataFile.Close()
			syncDataMsgChan <- true //启动重新同步

		default:
			logger.AsyncInfo(fmt.Sprintf("未识别的包, %#v", dataPackage))
		}

		logger.AsyncInfo("end 解包 ")
	}
}

//重连master
func (client *Client) reConnect() {
	_, err := net.ResolveTCPAddr("tcp", client.MasterAddress)
	CheckErr(err)

	connection, err := net.Dial("tcp", client.MasterAddress)
	CheckErr(err)

	client.Context.Connection = connection
	client.Context.LastActiveTs = time.Now().Unix()
}

//发送备份数据仓库的reqeust
func (client *Client) sendSyncDatabaseRequest(msgChan chan bool) {

	for {
		<- msgChan  //等待同步消息启动
		time.Sleep(10 * time.Second)

		//获取数据的请求包
		requestDataPackage := NewBackupPackage(ACTION_SYNC_DATA)
		requestDataPackage.encodeData(intToBytes(int(time.Now().Unix())))

		//logger.AsyncInfo(requestDataPackage)
		num, err := client.Context.writePackage(requestDataPackage)
		logger.AsyncInfo(fmt.Sprintf("发起数据同步请求:%#v字节 ,error: %#v", num, err))
	}

}

//发送心跳包
func (client *Client) sendHeartBeat() {

	for {

		pingPacakge := NewBackupPackage(ACTION_PING)
		pingPacakge.encodeData(intToBytes(int(time.Now().Unix())))
		//binary.Write(connection, binary.BigEndian, byte(ACTION_PING))
		//binary.Write(connection, binary.BigEndian, int32(4))
		//binary.Write(connection, binary.BigEndian, int32(time.Now().Unix()))
		//connection.Write(ACTION_PING)

		num, err := client.Context.writePackage(pingPacakge)
		logger.AsyncInfo(fmt.Sprintf("发起心跳包: send beat, %#v, %#v", num, err))

		time.Sleep(5 * time.Second)
	}
}