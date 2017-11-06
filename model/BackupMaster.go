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
	"sync"
	//"context"
	"os"
	"path"
)

//var	contextList *list.List

const (
	STATUS_NULL = 0x00
	STATUS_NEW = 0x01
	STATUS_FINISH = 0xFF

	TIME_FORMAT = "2006-01-02 15:04:05"
)

type Context struct {
	Connection net.Conn
	LastActiveTs int64 //最近一次活跃的时间戳
	Lock *sync.Mutex
}

//往socket中写数据
func (context *Context) writePackage(dataPackage *BackupPackage) (n int, err error) {

	defer func() {
		context.updateAliveTs() //更新活跃时间
		context.Lock.Unlock()
	}()

	context.Lock.Lock()

	writer := bufio.NewWriter(context.Connection)
	n, err = writer.Write(dataPackage.getHeader())
	//n, err = context.Connection.Write(dataPackage.getHeader())
	if err != nil {
		return n, err
	}

	//n, err = context.Connection.Write(dataPackage.Data)
	n, err = writer.Write(dataPackage.Data)
	if err != nil {
		return n, err
	}

	writer.Flush()

	return n,err
}

func (context *Context) updateAliveTs() {
	context.LastActiveTs = time.Now().Unix()
}

type MasterServer struct{
	ContextList *list.List
}

var masterServer *MasterServer
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
		lock := new(sync.Mutex)
		var context = &Context{connection, now, lock}

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

func (masterServer *MasterServer) handleConnection(context *Context) {
	defer func() {
		context.Connection.Close()

		err := recover()
		if err != nil {
			logger.AsyncInfo(err)
		}
	}()

	for {

		dataPackage := GetDecodedPackageData(context.Connection)
		context.LastActiveTs = time.Now().Unix()

		masterServer.handleAction(context, dataPackage)
	}
}

func (masterServer *MasterServer) handleAction(context *Context, dataPacakge *BackupPackage) {

	if dataPacakge.DataLength < 0 {
		panic("数据包长度少于0")
	}

	logger.AsyncInfo("MasterServer, handleAction" + fmt.Sprintf("dataPacakge:%#v", dataPacakge))

	switch dataPacakge.ActionType {
	case ACTION_PING:
		dataPackage  := NewBackupPackage(ACTION_PING)
		dataPackage.encodeData(int32ToBytes(int(context.LastActiveTs)))
		n, err :=context.writePackage(dataPackage)
		logger.AsyncInfo(fmt.Sprintf("心跳包: %#v, $#v", n, err))
		checkErr(err)
		break

	case ACTION_SYNC_DATA:
		logger.AsyncInfo("开始备份数据\t" + time.Now().Format(TIME_FORMAT) )

		// start 复制临时文件
		srcFile, err := os.Open(GetApplication().ConfigData.Bolt.FilePath)
		defer srcFile.Close()
		checkErr(err)

		destFilePath := path.Join(path.Dir(GetApplication().ConfigData.Bolt.FilePath), fmt.Sprintf("%d_%s_%s", os.Getpid(), MyMd5(context.Connection.RemoteAddr()), time.Now().Format("2006010215")))
		logger.AsyncInfo("临时文件路径:" + destFilePath)
		destFile, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE, 0644)
		defer os.Remove(destFilePath) //同步完成删除临时文件

		_, err = io.Copy(destFile, srcFile)
		checkErr(err)
		destFile.Close()

		//end 复制临时文件

		destFile, err = os.Open(destFilePath)
		checkErr(err)

		buffer := make([]byte, 1024)
		var isChunk bool = false
		var totalBytes int64 = 0;

		for {
			n, err := destFile.Read(buffer)
			if n <= 0 || (err != nil  && err != io.EOF) {
				logger.AsyncInfo(fmt.Sprintf("读文件内容异常, %d,  %#v", n, err))
				break
			}

			var dataPackage *BackupPackage

			if isChunk {
				dataPackage = NewBackupPackage(ACTION_CHUNK_DATA)
			} else {
				dataPackage = NewBackupPackage(ACTION_SYNC_DATA)
				isChunk = true
			}

			dataPackage.encodeData(buffer[0:n])
			_, err = context.writePackage(dataPackage)
			checkErr(err)

			totalBytes += int64(n)

			if n < 1024 {
				break
			}
		}
		logger.AsyncInfo(fmt.Sprintf("end备份数据\t%#v, %#v", time.Now().Format(TIME_FORMAT), totalBytes))
		break

	default:
		logger.AsyncInfo("不识别的action")
	}

	return
}

//处理新 connection  废弃了
func (masterServer *MasterServer) handleConnectionOld(context *Context) {
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
		logger.AsyncInfo("current status: " + strconv.Itoa(status))
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
				err = masterServer.handleActionOld(context, curAction, socketio, dataLength)

				dataLength = 0
				status = STATUS_NULL

				if err != nil {
					logger.AsyncInfo("MasterServer, handleAction" + fmt.Sprintf("%#v", err))
				}
		}
	}
}

//处理请求 废弃了
func (masterServer *MasterServer)handleActionOld(context *Context, action byte, socketio *bufio.ReadWriter, dataLength int) error {
	if dataLength < 0 {
		return errors.New("数据包长度少于0")
	}

	logger.AsyncInfo("MasterServer, handleAction" + fmt.Sprintf("action:%#v, dataLength:%#v", action, dataLength))

	if dataLength > 10000000 {
		socketio.Discard(dataLength)
		panic("数据包长度超过10M, 不允许")
	}

	context.updateAliveTs()

	switch action {
		case ACTION_PING:
			socketio.Discard(dataLength)

			dataPackage  := NewBackupPackage(ACTION_PING)
			dataPackage.encodeData(int32ToBytes(int(context.LastActiveTs)))

			n, err :=context.writePackage(dataPackage)
			//context.write()
			//socketio.Write(int32ToBytes(int(context.LastActiveTs))) // 转成package形式
			logger.AsyncInfo(fmt.Sprintf("ping action: %#v, $#v", n, err))
			break

		case ACTION_SYNC_DATA:
			socketio.Discard(dataLength)
			logger.AsyncInfo("开始备份数据\t" + time.Now().Format(TIME_FORMAT) )

			dataPackage := NewBackupPackage(ACTION_SYNC_DATA)
			dataPackage.encodeData([]byte("this is data from server, " + time.Now().Format(TIME_FORMAT) + "\n"))

			n, err := context.writePackage(dataPackage)

			//context.Connection.Write()) // 转成package形式
			//socketio.Flush()
			//binary.Write(socketio, binary.BigEndian, []byte("this is data from server\n"))
			logger.AsyncInfo(fmt.Sprintf("end备份数据\t%#v, %#v, %#v", time.Now().Format(TIME_FORMAT), n, err ))
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
