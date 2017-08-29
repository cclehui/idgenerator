package jsonApi

import (
	"github.com/gin-gonic/gin"
)

func Success(context *gin.Context, data gin.H, args ...int) {

	result := gin.H{
		"status":    "success",
		"message":   "",
		"errorCode": 0,
		"data":      data,
	}

	httpCode := 200

	if args != nil && len(args) > 0 {
		httpCode = args[0]
	}

	context.JSON(httpCode, result)
}

func Fail(context *gin.Context, message string, errorCode int, args ...int) {

	result := gin.H{
		"status":    "fail",
		"message":   message,
		"errorCode": errorCode,
		"data":      gin.H{},
	}

	httpCode := 200

	if args != nil && len(args) > 0 {
		httpCode = args[0]
	}

	context.JSON(httpCode, result)
}
