package model

import (
	"net"
	"time"
	"encoding/binary"
	"bufio"
	"idGenerator/model/logger"
	//"strconv"
	"fmt"
)

//var	contextList *list.List

type Client struct {
	Context *Context
}

var client *Client

//启动client 备份
func StartClientBackUp(masterAddress string) {
	_, err := net.ResolveTCPAddr("tcp", masterAddress)
	CheckErr(err)

	connection, err := net.Dial("tcp", masterAddress)
	CheckErr(err)

	now := time.Now().Unix()
	var context = &Context{connection, now}

	if client == nil {
		client = &Client{context}
	}

	//go client.sendHeartBeat()

	//备份数据库
	go client.syncDatabase()
}

//备份数据仓库
func (client *Client) syncDatabase() {

	for {
		time.Sleep(5 * time.Second)

		logger.AsyncInfo("开始同步数据")
		connection := client.Context.Connection

		reader := bufio.NewReader(connection)
		//writer := bufio.NewWriter(connection)

		//获取数据的请求包
		synDataPackage := make([]byte, 9)
		synDataPackage[0] = ACTION_SYNC_DATA
		synDataPackage[1] = 0
		synDataPackage[2] = 0
		synDataPackage[3] = 0
		synDataPackage[4] = 4
		//var synDataPackage = [9]byte{ACTION_SYNC_DATA,0,0,0,4}
		now := time.Now().Unix()
		binary.LittleEndian.PutUint32(synDataPackage[5:], uint32(now))

		logger.AsyncInfo(synDataPackage)
		//err := binary.Write(connection, binary.BigEndian, synDataPackage)
		//num, err := writer.Write(synDataPackage)
		num, err := connection.Write(synDataPackage)

		logger.AsyncInfo(fmt.Sprintf("写入:%#v字节 ,error: %#v", num, err))

		result,_,err := reader.ReadLine()
		logger.AsyncInfo(fmt.Sprintf("返回结果:%s, error:%#v" , result, err))

		//go func() {
		//
		//
		//}()
		logger.AsyncInfo("同步数据 end")

	}

}

//发送心跳包
func (client *Client) sendHeartBeat() {
	for {
		connection := client.Context.Connection
		//pingData = make([]byte, 9)
		//action = G
		binary.Write(connection, binary.BigEndian, byte(ACTION_PING))
		binary.Write(connection, binary.BigEndian, int32(4))
		binary.Write(connection, binary.BigEndian, int32(time.Now().Unix()))
		//connection.Write(ACTION_PING)

		time.Sleep(5 * time.Second)
	}
}