package main

import (
	_"gateway/config"
	"gateway/web"
	"gateway/mylog"
	_"gateway/server"
)

func main() {
	mylog.LOG.I("Gateway Server Start!")
	web.RunMartini()
}
