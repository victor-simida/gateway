package mylog

import (
	"os"
	"gateway/Godeps/_workspace/src/github.com/lixuanhao/fileLogger"
)

var (
	LOG *fileLogger.FileLogger
)

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func SetupLogger(appName string) {
	var path string

	path = "./logs"

	if !isExist(path) {
		os.Mkdir(path, 0755)
	}
	LOG = fileLogger.NewDailyLogger(path, appName+".log", "", fileLogger.DEFAULT_LOG_SCAN, fileLogger.DEFAULT_LOG_SEQ)
}

func init (){
	SetupLogger("gateway")
}
