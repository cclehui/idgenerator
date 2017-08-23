package logger

import (
    "log"
    "os"
    "time"
)

type MyLogger struct {
    Logger *log.Logger
    ChannelInfo chan string
}

var loggerInstance *MyLogger = nil;

//获取logger实例 单例模式
func GetLogger() *MyLogger{

    if (loggerInstance == nil) {
        loggerInstance = new(MyLogger);
        loggerInstance.Logger = log.New(os.Stdout, "[cclehui]\t", log.Ldate | log.Ltime)
        loggerInstance.ChannelInfo = make(chan string, 1024)

        go func() {
            select {
                case logData := <-loggerInstance.ChannelInfo : //info log 的channel
                    loggerInstance.Logger.Println(logData)
            }
        }()
    }

    return loggerInstance;
}


//异步写Log
func AsyncInfo(logStr string) {
    myLogger := GetLogger()

    timeOut := make(chan bool)

    go func() {
        select {
            case myLogger.ChannelInfo <- logStr:
                return
            case <-timeOut: //0.5s超时
                return
        }
    } ()

    //超时机制
    go func() {
        time.Sleep(500 * time.Millisecond)
        timeOut <- true
    } ()
}

//同步写Log
func Printf(format string, v ...interface{}) {
    myLogger := GetLogger();
    myLogger.Logger.Printf(format, v);
}
