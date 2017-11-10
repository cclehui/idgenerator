package main

//import  idGenerator "idGenerator/model"

import(
    "os"
	"encoding/binary"
	"bytes"
	"time"
	"idGenerator/model"
	"fmt"
	"idGenerator/model/logger"
	"strconv"
)

func bytesToInt32(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return int32(tmp)
}

func CheckErr(err interface{}) {
	if err != nil {
		panic(err)
	}
}

func doAction() {

	go func() {
		defer func() {

			err := recover()
			logger.AsyncInfo(fmt.Sprintf("self function , %#v", err))
		}()
		time.Sleep(4 * time.Second)
		panic("异常222222222222222")
	}()

	panic("异常111111111111：" + strconv.Itoa(int(time.Now().Unix())))
}



//每个业务对应一个 key 全局唯一
//var idWorkerMap = make(map[int]*idGenerator.IdWorker)
//var idWorkerMap = cmap.New();

func main() {


	channelRedo := make(chan bool)
	for {
		defer func() {
			err := recover()
			logger.AsyncInfo(fmt.Sprintf("for xdddddddddddd, %#v", err))
		}()

		go func() {
			defer func() {
				err := recover()
				logger.AsyncInfo(fmt.Sprintf("主从同步异常, %#v", err))

				channelRedo <- true
			}()

			logger.AsyncInfo("启动主从同步操作")
			doAction()



		}()
		<- channelRedo
		time.Sleep(2 * time.Second)
	}

	data := make(map[string]string)
	data["md5"] = model.CaculteFileMd5("./data/bolt_kv.db.backup")
	data["ts"] = time.Now().Format(model.TIME_FORMAT)
	data["xxx"] = model.CaculteFileMd5("./README.md")

	fmt.Println(data)

	//encodedData, _ := json.Marshal(data)
	//
	//var decodeData map[string]string
	//
	//json.Unmarshal(encodedData, &decodeData)
	//
	//fmt.Println(data)
	//fmt.Println(decodeData)
	//fmt.Println(decodeData["md5"])
	//
	//if decodeData["md5"] == "d41d8cd98f00b204e9800998ecf8427e" {
	//	fmt.Println("xxxxxxxxxxxxxxx")
	//}
	//重新以append 方式打开文件
	//backupDataFile, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	//defer backupDataFile.Close()
	//CheckErr(err)

	//data = []byte("yyyyyyyyyyyyyyyyyyy\n")
	//
	//backupDataFile.Write(data)

	//data := []byte{0x0,0x4,0x0,0x0}
	//
	//fmt.Printf("%d\n", bytesToInt32(data))
	//
	//temp := []byte{0x0, 0x0, 0x1, 0x2e};
	////length := bytesToInt32(temp)
	//fmt.Printf("%#v, %#v\n", temp, temp[0:2])

	//fmt.Printf("aaa:%#v", reflect.TypeOf(model.ACTION_SYNC_DATA).Name())
	os.Exit(0)

	//fmt.Println(time.Now().Format("2006-01-02 12:04:05"))
	//
    //var synDataPackage = [9]byte{2,0,0,0,4}
	//
    //now := time.Now().Unix()
	//
    //temp := make([]byte, 4)
    ////temp := synDataPackage[5:]
	//
    //binary.LittleEndian.PutUint32(synDataPackage[5:], uint32(now))
    //binary.LittleEndian.PutUint32(temp, uint32(now))
	//
    ////synDataPackage[5] = byte(now >> 3)
    ////synDataPackage[6] = byte(now >> 2)
    ////synDataPackage[7] = byte(now >> 1)
    ////synDataPackage[8] = byte(now)
	//
    //fmt.Println(synDataPackage);
    //fmt.Println(temp);
    //fmt.Println(binary.LittleEndian.Uint32(temp));
    //fmt.Println(now);
    //fmt.Println(binary.LittleEndian.Uint32(synDataPackage[5:]));
	//
    //os.Exit(0)

}
