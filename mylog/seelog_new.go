package mylog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gateway/Godeps/_workspace/src/github.com/cihub/seelog"
	"net/http"
	"runtime"
)

var LOG iLog

type iLog interface {
	Debug(format string, params ...interface{})
	Info(format string, params ...interface{})
	Warn(format string, params ...interface{}) error
	Error(format string, params ...interface{}) error
	Critical(format string, params ...interface{}) error

	I(format string, params ...interface{})
	W(format string, params ...interface{}) error
	E(format string, params ...interface{}) error
}

type logger struct {
	SrvName     string
	CfgFilePath string
	EnableLogMonitor	bool
}

func SeelogInit(srvname string, logCfgPath string, enableLogMonitor bool, logMonitorUrl string) {
	lg, err := seelog.LoggerFromConfigAsFile(logCfgPath)
	if err != nil {
		panic(err)
	}
	lg.SetAdditionalStackDepth(1)
	seelog.ReplaceLogger(lg)

	LOG = &logger{
		SrvName: srvname,
		EnableLogMonitor: enableLogMonitor,
	}

	ChanSendToLogMonitor = make(chan LogInfo_t, 1024)
	LogMonitorUrl = logMonitorUrl
	go sendToLogMonitorLoop()
	return
}

func (self *logger) Close() {
	seelog.Flush()
}

func (self *logger) Debug(format string, args ...interface{}) {
	seelog.Debugf(format, args...)
}

func (self *logger) Info(format string, args ...interface{}) {
	seelog.Infof(format, args...)
}

func (self *logger) Warn(format string, args ...interface{}) error {
	return seelog.Warnf(format, args...)
}

func (self *logger) Error(format string, args ...interface{}) error {
	if !self.EnableLogMonitor {
		return seelog.Errorf(format, args...)
	}

	seelogRet := seelog.Errorf(format, args...)
	_, file, line, _ := runtime.Caller(1)

	var loginfo LogInfo_t
	loginfo.SrvName = self.SrvName
	loginfo.Fileline = fmt.Sprintf("%s:%d", file, line)
	loginfo.Content = seelogRet.Error()
	AsyncSendToLogMonitor(loginfo)

	return seelogRet
}

func (self *logger) Critical(format string, args ...interface{}) error {
	return seelog.Criticalf(format, args...)
}

func (self *logger) I(format string, args ...interface{}) {
	seelog.Infof(format, args...)
}

func (self *logger) W(format string, args ...interface{}) error {
	return seelog.Warnf(format, args...)
}

func (self *logger) E(format string, args ...interface{}) error {
	if !self.EnableLogMonitor {
		return seelog.Errorf(format, args...)
	}

	seelogRet := seelog.Errorf(format, args...)
	_, file, line, _ := runtime.Caller(1)

	var loginfo LogInfo_t
	loginfo.SrvName = self.SrvName
	loginfo.Fileline = fmt.Sprintf("%s:%d", file, line)
	loginfo.Content = seelogRet.Error()
	AsyncSendToLogMonitor(loginfo)

	return seelogRet
}

//======================= 错误日志统一上报，以便监控 =======================
var ChanSendToLogMonitor chan LogInfo_t
var LogMonitorUrl string

type LogInfo_t struct {
	SrvName  string `json:"srvName"`
	Fileline string `json:"fileline"`
	Content  string `json:"content"`
}

func AsyncSendToLogMonitor(info LogInfo_t) {
	select {
	case ChanSendToLogMonitor <- info:
	default:
		fmt.Printf("AsyncSendToLogMonitor is full!!!\n")
	}
}

func sendToLogMonitorLoop() {
	for {
		info := <-ChanSendToLogMonitor
		logdata, _ := json.Marshal(&info)
		logPostJson(LogMonitorUrl, logdata)
	}
}

// 日志模块不能依赖proxy模块，否则会交叉引用，编译报错
func logPostJson(url string, data []byte) {
	method := "POST"
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Accept-Charset", "UTF-8")

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Print("[logPostJson-error] %v", err.Error())
		return
	}
	defer res.Body.Close()
}
