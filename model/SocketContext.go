package model

import (
	"time"
	"net"
	"sync"
	"bufio"
)

type Context struct {
	Connection net.Conn
	LastActiveTs int64 //最近一次活跃的时间戳
	Lock *sync.Mutex
	Reader *bufio.Reader
	Writer *bufio.Writer
}

//往socket中写数据
func (context *Context) writePackage(dataPackage *BackupPackage) (n int, err error) {

	defer func() {
		context.updateAliveTs() //更新活跃时间
		context.Lock.Unlock()
	}()

	context.Lock.Lock()

	//writer := bufio.NewWriter(context.Connection)
	writer := context.getWriter()
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

func (context *Context) getReader() *bufio.Reader {
	if context.Reader == nil {
		context.Reader = bufio.NewReader(context.Connection)
	}

	return context.Reader
}

func (context *Context) getWriter() *bufio.Writer {
	if context.Writer == nil {
		context.Writer = bufio.NewWriter(context.Connection)
	}

	return context.Writer
}