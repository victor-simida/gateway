package web

import (
	"gateway/Godeps/_workspace/src/github.com/go-martini/martini"
	"gateway/config"
	"net/http"
)

type regInfo struct {
	Method  string
	Uri     string
	Handler []martini.Handler
}

var _RegInfo []regInfo = make([]regInfo, 0, 50)

func RegisterHandler(method, uri string, handler ...martini.Handler) {
	info := regInfo{method, uri, handler}
	_RegInfo = append(_RegInfo, info)
}

func delRealIp (req *http.Request) {
	req.Header.Del("X-Real-Ip")
	req.Header.Del("X-Forwarded-For")
}

func RunMartini() {
	m := martini.Classic()
	m.Use(delRealIp)

	for _, info := range _RegInfo {
		switch info.Method {
		case "Get":
			m.Get(info.Uri, info.Handler...)
		case "Post":
			m.Post(info.Uri, info.Handler...)
		}
	}

	port := ""
	port = config.Settings.ServerPort

	m.RunOnAddr(`:` + port)
}
