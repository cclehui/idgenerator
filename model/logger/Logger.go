package logger

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"
)

type MyLogger struct {
	Logger      *log.Logger
	ChannelInfo chan string
}

var loggerInstance *MyLogger = nil

//获取logger实例 单例模式
func GetLogger() *MyLogger {

	if loggerInstance == nil {
		loggerInstance = new(MyLogger)
		loggerInstance.Logger = log.New(os.Stdout, "[cclehui]\t", log.Ldate|log.Ltime)
		loggerInstance.ChannelInfo = make(chan string, 1024)

		go func() {
			for {
				select {
				case logData := <-loggerInstance.ChannelInfo: //info log 的channel
					loggerInstance.Logger.Println(logData)
				}
			}
		}()
	}

	return loggerInstance
}

func AsyncDebug(logData interface{}) {


}

//异步写Log
func AsyncInfo(logData interface{}) {
	myLogger := GetLogger()

	var logStr string

	switch value := logData.(type) {
		case string:
			logStr = value
		default:
			logStr = fmt.Sprintf("%#v", logData)
	}

	select {
	case myLogger.ChannelInfo <- logStr:
		return
		//case <-timeOut: //0.5s超时
	case <- time.After(1000 * time.Millisecond):
		fmt.Println("write log time out")
		return
	}

	//timeOut := make(chan bool)

	//go func() {
	//	select {
	//	case myLogger.ChannelInfo <- logStr:
	//		return
	//	//case <-timeOut: //0.5s超时
	//	case <- time.After(1000 * time.Millisecond):
	//		fmt.Println("write log time out")
	//		return
	//	}
	//}()

	//超时机制
	//go func() {
	//	time.Sleep(500 * time.Millisecond)
	//	timeOut <- true
	//}()
}

//同步写Log
func Printf(format string, v ...interface{}) {
	myLogger := GetLogger()
	myLogger.Logger.Printf(format, v)
}

//log 中间件
var (
	green        = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white        = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow       = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red          = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue         = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta      = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan         = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
)

//写Log的 middleware
func LoggerHanderFunc() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()
		path := context.Request.URL.Path

		// Process request
		context.Next()

		end := time.Now()
		latency := end.Sub(start)

		clientIP := context.ClientIP()
		method := context.Request.Method
		statusCode := context.Writer.Status()
		var statusColor, methodColor string

		statusColor = colorForStatus(statusCode)
		methodColor = colorForMethod(method)

		comment := context.Errors.ByType(gin.ErrorTypePrivate).String()

		//logData := fmt.Sprintf("%v |%s %3d %s| %13v | %15s |%s  %s %-7s %s\n%s",
		logData := fmt.Sprintf("%s %3d %s| %13v | %15s |%s  %s %-7s %s %s",
			//end.Format("2006/01/02 - 15:04:05"),
			statusColor, statusCode, reset,
			latency,
			clientIP,
			methodColor, method, reset,
			path,
			comment,
		)

		//异步写Log
		AsyncInfo(logData)
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}
