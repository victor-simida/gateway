package main

import (
	_"gateway/config"
	"gateway/web"
	"gateway/mylog"
	"gateway/server"
	"flag"
)

var SeelogCfgPath = flag.String("seelogCfgPath", "./config/seelog.xml", "seelog config file")

func main() {
	// ?????
	mylog.SeelogInit("gateway", *SeelogCfgPath, false, "")
	server.Init()

	mylog.LOG.I("Gateway Server Start!")
	web.RunMartini()
}
