package persistent

import (
	"net"
	"time"
	"container/list"
	"idGenerator/model/logger"
)

var	contextList *list.List

type Context struct {
	Connection net.Conn
	LastActiveTs int64 //最近一次活跃的时间戳
}

func StartMasterServer(serverAddress string) {
	tcpAddress, err := net.ResolveTCPAddr("tcp", serverAddress)
	CheckErr(err)

	listener, err := net.Listen("tcp", tcpAddress)
	CheckErr(err)

	defer func() {
		listener.Close()
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
	}

	var status = STATUS_NEW //状态机
	var dataLength int = 0
	var err error
	var curAction byte

	socketio := buffio.NewReadWriter(context.Connection, context.Connection);

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
				if err != nil {
					logger.AsyncInfo("MasterServer, handleAction" + fmt.Sprintf("%#v", err))
				}
				dataLength = 0
				status = STATUS_NULL
		}
	}
}

//处理请求
func handleAction(context *Context, action byte, socketio buffio.ReadWriter, dataLength int) error {
	if length < 0 {
		return errors.New("数据包长度少于0")
	}

	if length > 10000000 {
		panic("数据包长度超过10M, 不允许")
	}

	switch action {
		case ACTION_PING:
			socketio.Discard(4)
			context.LastActiveTs = time.Now().Unix()
			socketio.Write(int32ToBytes(context.LastActiveTs)) // 转成package形式
			break
	}


}

//获取数据包的长度
func getDataLength(socketio buffio.Reader) (int, error) {
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


func CheckErr(err interface{}) {
	if err != nil {
		panic(err)
	}
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
