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
)

var	contextList *list.List

type Context struct {
	Connection net.Conn
	LastActiveTs int64 //最近一次活跃的时间戳
}

func StartMasterServer(serverAddress string) {
	_, err := net.ResolveTCPAddr("tcp", serverAddress)
	CheckErr(err)

	listener, err := net.Listen("tcp", serverAddress)
	CheckErr(err)

	defer func() {
		listener.Close()
	}()

	if contextList == nil {
		contextList = list.New()
	}

	// 开启一个子 grountine 来遍历 contextList cclehui_todo

	for {
		connection, err := listener.Accept()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		now := time.Now().Unix()
		var context = &Context{connection, now}

		logger.AsyncInfo("new connection:" + fmt.Sprintf("%#v", connection))

		contextList.PushBack(context) //放入全局context list中

		go handleConnection(context)

	}
}

const (
	STATUS_NULL = 0x00
	STATUS_NEW = 0x01
	STATUS_FINISH = 0xFF

	//action 类型
	ACTION_PING = 0x01
	ACTION_SYNC_DATA = 0x02 //同步数据
)

func handleConnection(context *Context) {
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
				err = handleAction(context, curAction, socketio, dataLength)

				dataLength = 0
				status = STATUS_NULL

				if err != nil {
					logger.AsyncInfo("MasterServer, handleAction" + fmt.Sprintf("%#v", err))
				}
		}
	}
}

//处理请求
func handleAction(context *Context, action byte, socketio *bufio.ReadWriter, dataLength int) error {
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

	switch action {
		case ACTION_PING:
			socketio.Discard(dataLength)
			context.LastActiveTs = time.Now().Unix()
			socketio.Write(int32ToBytes(int(context.LastActiveTs))) // 转成package形式
			break
		case ACTION_SYNC_DATA:
			socketio.Discard(dataLength)
			context.LastActiveTs = time.Now().Unix()
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
