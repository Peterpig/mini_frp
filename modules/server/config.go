package server

import (
	"fmt"
	"strconv"

	"github.com/vaughan0/go-ini"
)

var (
	BindAddr         string = "0.0.0.0"
	BindPort         int64  = 7000
	LogFile          string = "./frps.log"
	LogLevel         string = "warn"
	LogWay           string = "console"
	HeartBeatTimeout int64  = 30
	UserConnTimeout  int64  = 100
)

var ProxyServers map[string]*ProxyServer = make(map[string]*ProxyServer)

func LoadConf(confFile string) (err error) {
	var tmpStr string
	var ok bool

	conf, err := ini.LoadFile(confFile)
	if err != nil {
		return err
	}

	tmpStr, ok = conf.Get("common", "bind_addr")
	if ok {
		BindAddr = tmpStr
	}

	tmpStr, ok = conf.Get("common", "bind_port")
	if ok {
		BindPort, err = strconv.ParseInt(tmpStr, 10, 64)
		if err != nil {
			return err
		}
	}

	tmpStr, ok = conf.Get("common", "log_file")
	if ok {
		LogFile = tmpStr
	}

	tmpStr, ok = conf.Get("common", "log_level")
	if ok {
		LogLevel = tmpStr
	}

	tmpStr, ok = conf.Get("common", "log_way")
	if ok {
		LogWay = tmpStr
	}

	for name, section := range conf {
		if name == "common" {
			continue
		}

		proxyServer := &ProxyServer{}
		proxyServer.Name = name
		if proxyServer.Passwd, ok = section["passwd"]; !ok {
			return fmt.Errorf("parse ini file error: proxy [%s] no passwd found", name)
		}

		if proxyServer.BindAddr, ok = section["bind_addr"]; !ok {
			proxyServer.BindAddr = "0.0.0.0"
		}

		portStr, ok := section["bind_port"]
		if !ok {
			return fmt.Errorf("parse ini file error: proxy [%s] no bind_port found", name)
		}

		proxyServer.BindPort, err = strconv.ParseInt(portStr, 10, 64)
		if err != nil {
			return fmt.Errorf("parse ini file error: proxy [%s] port error: %v", name, err)
		}

		proxyServer.Init()
		ProxyServers[proxyServer.Name] = proxyServer
	}

	if len(ProxyServers) == 0 {
		return fmt.Errorf("parse ini filer error: no proxy client found")
	}

	return
}
