package main

//import  idGenerator "idGenerator/model"

import(
	"fmt"
    "os"
	//"encoding/binary"
	"time"
	"encoding/binary"
	"bytes"
)

func bytesToInt32(b []byte) uint32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return uint32(tmp)
}


//每个业务对应一个 key 全局唯一
//var idWorkerMap = make(map[int]*idGenerator.IdWorker)
//var idWorkerMap = cmap.New();

func main() {

	fmt.Println(time.Now().Format("2006-01-02 12:04:05"))

    var synDataPackage = [9]byte{2,0,0,0,4}

    now := time.Now().Unix()

    temp := make([]byte, 4)
    //temp := synDataPackage[5:]

    binary.LittleEndian.PutUint32(synDataPackage[5:], uint32(now))
    binary.LittleEndian.PutUint32(temp, uint32(now))

    //synDataPackage[5] = byte(now >> 3)
    //synDataPackage[6] = byte(now >> 2)
    //synDataPackage[7] = byte(now >> 1)
    //synDataPackage[8] = byte(now)

    fmt.Println(synDataPackage);
    fmt.Println(temp);
    fmt.Println(binary.LittleEndian.Uint32(temp));
    fmt.Println(now);
    fmt.Println(binary.LittleEndian.Uint32(synDataPackage[5:]));

    os.Exit(0)

}
