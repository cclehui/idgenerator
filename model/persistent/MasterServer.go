package persistent

import (
	"net"
	"time"
	"container/list"
)

var	ContextList *list.List

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

	for {
		connection, err := listener.Accept()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		now := time.Now().Unix()
		var context = &Context{connection, now}

		ContextList.PushBack(context) //放入全局context list中

		go handleConnection(context)

	}
}

func handleConnection(context *Context) {
	//cclehui_todo
}

func CheckErr(err interface{}) {
	if err != nil {
		panic(err)
	}
}
