package web

import (
	"gateway/Godeps/_workspace/src/github.com/go-martini/martini"
	"gateway/Godeps/_workspace/src/github.com/martini-contrib/gzip"
	"gateway/config"
	"net/http"
)

type regInfo struct {
	Uri     string
	Handler []martini.Handler
}

var _RegInfo []regInfo = make([]regInfo, 0, 50)

func RegisterHandler(uri string, handler ...martini.Handler) {
	info := regInfo{uri, handler}
	_RegInfo = append(_RegInfo, info)
}

func delRealIp (req *http.Request) {
	req.Header.Del("X-Real-Ip")
	req.Header.Del("X-Forwarded-For")
}

func  RunMartini() {
	m := martini.Classic()
	m.Use(delRealIp)
	m.Use(gzip.All())
	for _, info := range _RegInfo {
		m.Get(info.Uri, info.Handler...)
		m.Post(info.Uri, info.Handler...)
	}

	port := ""
	port = config.Settings.ServerPort

	m.RunOnAddr(`:` + port)
}
