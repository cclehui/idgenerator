package model

import (
	"io"
	"encoding/gob"
	"bufio"
	"net/rpc"
	"idGenerator/model/logger"
	"fmt"
)

//增加一个超时 timeout 处理 cclehui_todo

type GobServerCodec struct {
	rwc    *Context
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *GobServerCodec) ReadRequestHeader(r *rpc.Request) error {

	if c.closed {
		return io.EOF  //退出上层的for 循环
	}

	c.rwc.updateAliveTs() //更新活跃时间

	return c.dec.Decode(r)
}

func (c *GobServerCodec) ReadRequestBody(body interface{}) error {
	if c.closed {
		return io.EOF  //退出上层的for 循环
	}

	c.rwc.updateAliveTs() //更新活跃时间

	return c.dec.Decode(body)
}

func (c *GobServerCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	c.rwc.updateAliveTs()

	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Gob couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			logger.AsyncInfo(fmt.Sprintf("rpc: gob error encoding response:%#v", err))
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a gob problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			logger.AsyncInfo(fmt.Sprintf("rpc: gob error encoding body:%#v", err))
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *GobServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Connection.Close()
}


type GobClientCodec struct {
	rwc    *Context
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func (c *GobClientCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *GobClientCodec) ReadResponseHeader(r *rpc.Response) error {
	return c.dec.Decode(r)
}

func (c *GobClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobClientCodec) Close() error {
	return c.rwc.Connection.Close()
}