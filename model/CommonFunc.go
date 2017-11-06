package model

import (
	"crypto/md5"
	"fmt"
)

func MyMd5(data interface{}) string {

	result := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%#v", data))))

	return result
}
