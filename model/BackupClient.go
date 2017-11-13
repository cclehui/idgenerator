package model

import (
	"net"
	"time"
	"idGenerator/model/logger"
	"fmt"
	"sync"
	"os"
	"encoding/json"
)

//var	contextList *list.List

type Client struct {
	Context *Context
	MasterAddress string
}

//var client *Client

//获取新连接
func NewClient(masterAddress string) *Client {
	tcpAddress, err := net.ResolveTCPAddr("tcp", masterAddress)
	CheckErr(err)

	//connection, err := net.Dial("tcp", masterAddress)
	connection, err := net.DialTCP("tcp", nil, tcpAddress)
	CheckErr(err)

	now := time.Now().Unix()
	lock := new(sync.Mutex)
	var context = &Context{connection, now,lock,nil,nil}

	client := &Client{context, masterAddress}

	return client

}

//启动client 备份
func (client *Client) StartClientBackUp()  {
	//备份数据库
	channelRedo := make(chan bool)
	for {
		go func() {
			defer func() {
				err := recover()
				logger.AsyncInfo(fmt.Sprintf("主从同步异常, %#v", err))

				channelRedo <- true
			}()

			logger.AsyncInfo("启动主从同步操作")
			client.doAction()



		}()
		<- channelRedo
		time.Sleep(2 * time.Second)
	}
}

func (client *Client) doAction() {
	defer func() {
		err := recover()
		if err != nil {
			logger.AsyncInfo(fmt.Sprintf("doAction error : %#v",err))

			if _, ok := err.(*net.OpError); ok {
				err = client.Context.Connection.Close()
				logger.AsyncInfo(fmt.Sprintf("重连master : %#v",err))
				client.reConnect()//尝试重连
			}
		}
	}()

	//发送心跳包
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
		//logger.AsyncInfo(fmt.Sprintf("开始解包: count:%d, action: %#v, length:%d", count, dataPackage.ActionType, dataPackage.DataLength))

		switch dataPackage.ActionType {
		case ACTION_PING:
			//logger.AsyncInfo("心跳包返回")
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
			if backupDataFile != nil {
				logger.AsyncInfo(fmt.Sprintf("同步完成， 共同步数据 : %d bytes", totalSize))
				backupDataFile.Close()
				totalSize = 0
			}
			syncDataMsgChan <- true //启动重新同步

		default:
			logger.AsyncInfo(fmt.Sprintf("未识别的包, %#v", dataPackage))
		}

		//logger.AsyncInfo("end 解包 ")
	}
}

//重连master
func (client *Client) reConnect() {
	_, err := net.ResolveTCPAddr("tcp", client.MasterAddress)
	CheckErr(err)

	connection, err := net.Dial("tcp", client.MasterAddress)
	CheckErr(err)

	now := time.Now().Unix()
	lock := new(sync.Mutex)
	var context = &Context{connection, now,lock,nil,nil}

	client.Context = context
}

//发送备份数据仓库的reqeust
func (client *Client) sendSyncDatabaseRequest(msgChan chan bool) {
	defer func() {
		err := recover()
		logger.AsyncInfo(fmt.Sprintf("sendSyncDatabaseRequest error : %#v",err))
	}()

	for {
		<- msgChan  //等待同步消息启动
		time.Sleep(2 * time.Second)

		data := make(map[string]string)
		data["md5"] = CaculteFileMd5(GetApplication().ConfigData.Bolt.FilePath)
		data["ts"] = time.Now().Format(TIME_FORMAT)

		encodedData, _ := json.Marshal(data)

		//获取数据的请求包
		requestDataPackage := NewBackupPackage(ACTION_SYNC_DATA)
		//requestDataPackage.encodeData(intToBytes(int(time.Now().Unix())))
		requestDataPackage.encodeData(encodedData)

		//logger.AsyncInfo(requestDataPackage)
		num, err := client.Context.writePackage(requestDataPackage)
		logger.AsyncInfo(fmt.Sprintf("发起数据同步请求:%#v字节 ,error: %#v", num, err))
	}
}

//发送心跳包
func (client *Client) sendHeartBeat() {
	defer func() {
		err := recover()
		logger.AsyncInfo(fmt.Sprintf("sendHeartBeat error : %#v",err))
	}()

	for {

		pingPacakge := NewBackupPackage(ACTION_PING)
		pingPacakge.encodeData(intToBytes(int(time.Now().Unix())))
		//binary.Write(connection, binary.BigEndian, byte(ACTION_PING))
		//binary.Write(connection, binary.BigEndian, int32(4))
		//binary.Write(connection, binary.BigEndian, int32(time.Now().Unix()))
		//connection.Write(ACTION_PING)

		num, err := client.Context.writePackage(pingPacakge)
		logger.AsyncInfo(fmt.Sprintf("发起心跳包: send beat, %#v, %#v", num, err))
		checkErr(err)

		time.Sleep(5 * time.Second)
	}
}