package model

import (
	"net"
	"time"
	//"encoding/binary"
	"idGenerator/model/logger"
	//"strconv"
	"fmt"
	"sync"
	//"bufio"
)

//var	contextList *list.List

type Client struct {
	Context *Context
	MasterAddress string
}

var client *Client

//启动client 备份
func StartClientBackUp(masterAddress string) {
	_, err := net.ResolveTCPAddr("tcp", masterAddress)
	CheckErr(err)

	connection, err := net.Dial("tcp", masterAddress)
	CheckErr(err)

	now := time.Now().Unix()
	lock := new(sync.Mutex)
	var context = &Context{connection, now,lock}

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
				client.reConnect()//尝试重连
			}
		}
	}()

	go client.sendHeartBeat()

	//发送数据备份的请求
	go client.sendSyncDatabaseRequest()

	//读数据
	for {
		logger.AsyncInfo("开始解包")
		connection := client.Context.Connection
		dataPackage := GetDecodedPackageData(connection)

		switch dataPackage.ActionType {
		case ACTION_PING:
			logger.AsyncInfo("心跳包返回")
		case ACTION_SYNC_DATA:
			logger.AsyncInfo(fmt.Sprintf("解包结果:%#v, length:%d, data:%#v", dataPackage.ActionType, dataPackage.DataLength, string(dataPackage.Data)))
		default:
			logger.AsyncInfo(fmt.Sprintf("未识别的包, %#v", dataPackage))
		}
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
func (client *Client) sendSyncDatabaseRequest() {

	for {
		time.Sleep(3 * time.Second)

		logger.AsyncInfo("开始同步数据")


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
		logger.AsyncInfo(fmt.Sprintf("heart beat: send beat, %#v, %#v", num, err))

		time.Sleep(5 * time.Second)
	}
}