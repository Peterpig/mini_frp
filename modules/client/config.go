package client

import (
	"fmt"
	"strconv"

	"github.com/vaughan0/go-ini"
)

var ClientProxys map[string]*ClientProxy = make(map[string]*ClientProxy)

var (
	ServerAddr        string = "0.0.0.0"
	ServerPort        int64  = 7000
	LogFile           string = "./frpc.log"
	LogLevel          string = "debug"
	LogWay            string = "console"
	HeartBeatTimeout  int64  = 30
	HeartBeatInterval int64  = 5
)

func LoadConf(confFile string) (err error) {
	var tmpStr string
	var ok bool

	conf, err := ini.LoadFile(confFile)
	if err != nil {
		return err
	}

	tmpStr, ok = conf.Get("common", "server_addr")
	if ok {
		ServerAddr = tmpStr
	}

	tmpStr, ok = conf.Get("common", "server_port")
	if ok {
		ServerPort, err = strconv.ParseInt(tmpStr, 10, 64)
		if err != nil {
			return err
		}
	}

	tmpStr, ok = conf.Get("common", "log_level")
	if ok {
		LogLevel = tmpStr
	}

	tmpStr, ok = conf.Get("common", "log_file")
	if ok {
		LogFile = tmpStr
	}

	tmpStr, ok = conf.Get("common", "log_way")
	if ok {
		LogWay = tmpStr
	}

	for name, section := range conf {
		if name == "common" {
			continue
		}

		clientProxy := &ClientProxy{}
		clientProxy.Name = name

		clientProxy.Passwd, ok = section["passwd"]
		if !ok {
			return fmt.Errorf("parse ini file error, proxy [%s] no passwd found", name)
		}

		portStr, ok := section["local_port"]
		if !ok {
			return fmt.Errorf("parse ini file error, proxy [%s] no local_port found", name)
		}

		if clientProxy.LocalPort, err = strconv.ParseInt(portStr, 10, 64); err != nil {
			return fmt.Errorf("parse ini file error, proxy [%s] local_port err: %v", name, err)
		}

		ClientProxys[name] = clientProxy
	}

	if len(ClientProxys) == 0 {
		return fmt.Errorf("parse ini file error: no proxy config found")
	}

	return
}
