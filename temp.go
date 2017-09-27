package main

import (
    "fmt"
    "errors"
    "bytes"
    "encoding/binary"
    "idGenerator/model/cmap"
);

type Application struct {
    idWorkerMap cmap.ConcurrentMap
}

var application Application;

func main() {
    //fmt.Println("xxxxxxxx");
    //fmt.Println(application);

	var data = 100
	var bytesData = intToBytes(data)

	fmt.Println(bytesData)
	fmt.Println(bytesToInt(bytesData))

}

//整形转换成字节  
func intToBytes(n int) []byte {
    bytesBuffer := bytes.NewBuffer([]byte{})
	temp := int32(n)
    binary.Write(bytesBuffer, binary.BigEndian, temp)
    return bytesBuffer.Bytes()
}

//字节转换成整形  
func bytesToInt(b []byte) int {
    bytesBuffer := bytes.NewBuffer(b)
    var tmp int32
    binary.Read(bytesBuffer, binary.BigEndian, &tmp)
    return int(tmp)
}

func test() (result int, err error) {

    defer func() {
        e := recover();

        if panicErr, ok := e.(error); ok {
            err = panicErr
            fmt.Printf("3333333333:%#v\n" , err)
        } else {
            //panic(e)
        }
    }()

    fmt.Println("11111111")
    panic(errors.New("eeeeeeeee"))
    fmt.Println("22222222")

    return 1, nil
}
