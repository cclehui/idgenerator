package model

import (
	"net"
	"time"
	"io"
	"fmt"
	"errors"
	"bytes"
	"encoding/binary"
	"bufio"
	"container/list"
	"idGenerator/model/logger"
	"math"
	"strconv"
)

//var	contextList *list.List

const (
	STATUS_NULL = 0x00
	STATUS_NEW = 0x01
	STATUS_FINISH = 0xFF

	//action 类型
	ACTION_PING = 0x01
	ACTION_SYNC_DATA = 0x02 //同步数据
)

type Context struct {
	Connection net.Conn
	LastActiveTs int64 //最近一次活跃的时间戳
}

type MasterServer struct{
	ContextList *list.List
}

type Client struct {
	Context *Context
}

var masterServer *MasterServer
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

	go client.sendHeartBeat()

	//备份数据库
	go client.syncDatabase()
}

//备份数据仓库
func (client *Client) syncDatabase() {

	for {
		time.Sleep(6 * time.Second)

		logger.AsyncInfo("开始同步数据")
		connection := client.Context.Connection

		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)

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
		num, err := writer.Write(synDataPackage)

		logger.AsyncInfo("已写入:\t" + strconv.Itoa(num))

		result,_,err := reader.ReadLine()

		logger.AsyncInfo(err)
		logger.AsyncInfo("返回结果:\t"  + string(result))

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

//启动master server
func StartMasterServer(serverAddress string) {
	_, err := net.ResolveTCPAddr("tcp", serverAddress)
	CheckErr(err)

	listener, err := net.Listen("tcp", serverAddress)
	CheckErr(err)

	defer func() {
		listener.Close()
	}()

	logger.AsyncInfo("start master server on :" + serverAddress)

	if masterServer == nil {
		masterServer = &MasterServer{list.New()}
	}

	// 开启一个子 grountine 来遍历 contextList cclehui_todo
	go masterServer.doConnectionAliveCheck()

	for {
		connection, err := listener.Accept()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		now := time.Now().Unix()
		var context = &Context{connection, now}

		logger.AsyncInfo("new connection:" + fmt.Sprintf("%#v", connection))

		masterServer.ContextList.PushBack(context) //放入全局context list中

		go masterServer.handleConnection(context)

	}
}

//连接活跃情况检查
func (masterServer *MasterServer) doConnectionAliveCheck() {
	for {
		maxUnActiveTs := int64(math.Max(float64(GetApplication().ConfigData.MaxUnActiveTs), 10.0))

		for item := masterServer.ContextList.Front(); item != nil; item = item.Next() {
			context, ok := item.Value.(*Context)
			if !ok {
				masterServer.ContextList.Remove(item)
			}

			now := time.Now().Unix()

			if now - context.LastActiveTs > maxUnActiveTs {
				context.Connection.Close()
				logger.AsyncInfo("超时关闭连接:" + fmt.Sprintf("now:%#v, connection%#v", now, context))
				masterServer.ContextList.Remove(item)
			}
		}

		time.Sleep(1 * time.Second)
	}
}


//处理新 connection
func (masterServer *MasterServer)handleConnection(context *Context) {
	//cclehui_todo

	defer func() {
		context.Connection.Close()

		err := recover()
		if err != nil {
			logger.AsyncInfo(err)
		}
	}()

	var status = STATUS_NULL //状态机
	var dataLength int = 0
	var err error
	var curAction byte

	ioReader := bufio.NewReader(context.Connection)
	ioWriter := bufio.NewWriter(context.Connection)
	socketio := bufio.NewReadWriter(ioReader, ioWriter);

	FORLABEL:
	for {
		switch status {
			case STATUS_NULL:
				curAction, err = socketio.ReadByte()
				if err != nil {
					if err == io.EOF {
						status = STATUS_NULL
						time.Sleep(1 * time.Second)
						logger.AsyncInfo("socket 无数据")
						break
					} else {
						break FORLABEL;
					}
				}

				logger.AsyncInfo("new action byte:" + fmt.Sprintf("%#v", curAction))

				if isNewAction(curAction) {
					dataLength, err = getDataLength(socketio)
					status = STATUS_NEW
					if err != nil {
						dataLength = 0
						status = STATUS_NULL
					}
				}

			case STATUS_NEW:
				err = masterServer.handleAction(context, curAction, socketio, dataLength)

				dataLength = 0
				status = STATUS_NULL

				if err != nil {
					logger.AsyncInfo("MasterServer, handleAction" + fmt.Sprintf("%#v", err))
				}
		}
	}
}

//处理请求
func (masterServer *MasterServer)handleAction(context *Context, action byte, socketio *bufio.ReadWriter, dataLength int) error {
	if dataLength < 0 {
		return errors.New("数据包长度少于0")
	}

	logger.AsyncInfo("MasterServer, handleAction" + fmt.Sprintf("action:%#v, dataLength:%#v", action, dataLength))

	if dataLength > 10000000 {
		socketio.Discard(dataLength)
		panic("数据包长度超过10M, 不允许")
	}

	//var buffer = make([]byte, 1024)
	//var data = make([]byte, 1024)

	context.LastActiveTs = time.Now().Unix()

	switch action {
		case ACTION_PING:
			socketio.Discard(dataLength)
			socketio.Write(int32ToBytes(int(context.LastActiveTs))) // 转成package形式
			break
		case ACTION_SYNC_DATA:
			socketio.Discard(dataLength)
			socketio.Write([]byte("this is data from server")) // 转成package形式
			break

		default:
			break
	}

	return nil
}

//获取数据包的长度
func getDataLength(socketio *bufio.ReadWriter) (int, error) {
	var byteSlice = make([]byte, 4)

	n, err := socketio.Read(byteSlice)
	if err != nil || n < 4 {
		return 0, errors.New("数据长度获取失败")
	}

	return bytesToInt32(byteSlice), nil
}

//是否是可识别的action
func isNewAction(action byte) bool {
	if action == ACTION_PING || action == ACTION_SYNC_DATA {
		return true
	}

	return false
}

//整形转换成字节  
func int32ToBytes(n int) []byte {
    bytesBuffer := bytes.NewBuffer([]byte{})
	tmp := int32(n)
    binary.Write(bytesBuffer, binary.BigEndian, tmp)
    return bytesBuffer.Bytes()
}

//字节转换成整形  
func bytesToInt32(b []byte) int {
    bytesBuffer := bytes.NewBuffer(b)
    var tmp int32
    binary.Read(bytesBuffer, binary.BigEndian, &tmp)
    return int(tmp)
}
