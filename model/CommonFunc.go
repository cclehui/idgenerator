package model

import (
	"crypto/md5"
	"fmt"
	"os"
	"io"
)

func MyMd5(data interface{}) string {

	result := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%#v", data))))

	return result
}

//计算文件的 md5
func CaculteFileMd5(filePath string) string {

	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}

	defer file.Close()

	md5Hash := md5.New()

	io.Copy(md5Hash, file)

	result := fmt.Sprintf("%x", md5.Sum(nil))

	return result
}
