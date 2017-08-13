package logger

import (
    "log"
    "os"
)

func GetLogger() *log.Logger{
    logger := log.New(os.Stdout, "cclehui_log\t", 1);

    return logger;
}

var loggerInstance *log.Logger = nil;

func Printf(format string, v ...interface{}) {
    if loggerInstance == nil {
        loggerInstance = GetLogger();
    }

    loggerInstance.Printf(format, v);
}
