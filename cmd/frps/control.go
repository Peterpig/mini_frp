package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mini_frp2/modules/conn"
	"mini_frp2/modules/consts"
	"mini_frp2/modules/msg"
	"mini_frp2/modules/server"
	"mini_frp2/utils/log"
	"time"
)

func ProcessControl(l *conn.Linster) {
	for {
		c, err := l.GetConn()
		if err != nil {
			return
		}
		log.Debug("Get one new conn, %v", c.GetRemoteAddr())
		go ProcessConn(c)
	}
}

func ProcessConn(c *conn.Conn) {
	content, err := c.ReadLine()
	if err != nil {
		log.Warn("Read data error: %v")
		return
	}

	log.Debug("Read data : %s", content)

	clientReq := &msg.RequestMsg{}
	clientResp := &msg.ResponseMsg{}

	if err := json.Unmarshal([]byte(content), &clientReq); err != nil {
		log.Error("Parse data err: %v", err)
		return
	}

	succ, info, needRes := check_proxy(clientReq, c)
	if !succ {
		clientResp.Code = 1
		clientResp.Msg = info
	}

	if needRes {
		buf, _ := json.Marshal(clientResp)
		err := c.Write(string(buf) + "\n")
		if err != nil {
			log.Error("ProxyName [%s], write error: %v", clientReq.ProxyName, err)
			return
		}
	} else {
		return
	}

	s := server.ProxyServers[clientReq.ProxyName]

	go heart_beat(s, c)

	serverReq := &msg.ResponseMsg{
		Code: consts.Working,
	}
	buf, _ := json.Marshal(serverReq)
	for {
		closeFlag := s.WaitUserConn()
		log.Info("closeFlag ==== %s", closeFlag)
		if closeFlag {
			log.Debug("ProxyName [%s], goroutine for dealing user conn is closed", s.Name)
			break
		}

		err := c.Write(string(buf) + "\n")
		if err != nil {
			log.Warn("ProxyName [%s], write to client err, %v, exit", s.Name, err)
			s.Close()
			return
		}

		log.Debug("ProxyName [%s] write to client to add work success", s.Name)
	}

	log.Info("ProxyName [%s], I'm dead!", s.Name)
}

func check_proxy(req *msg.RequestMsg, c *conn.Conn) (succ bool, info string, needRes bool) {
	succ = false
	needRes = true

	s, ok := server.ProxyServers[req.ProxyName]
	if !ok {
		info = fmt.Sprintf("ProxyName [%s], is not exist", req.ProxyName)
		return
	}

	if s.Passwd != req.Passwd {
		info = fmt.Sprintf("ProxyName [%s], password is not correct", req.ProxyName)
		return
	}

	if req.Type == consts.ClientConn {
		if s.Status != consts.Idle {
			log.Info("s.status: %s", s.Status)
			info = fmt.Sprintf("ProxyName [%s], already in use", req.ProxyName)
			log.Warn(info)
			return
		}
		err := s.Start()
		if err != nil {
			info = fmt.Sprintf("ProxyName [%s], start error: %v", req.ProxyName, err)
			return
		}

		log.Info("ProxyName [%s], start proxy success ", req.ProxyName)
	} else if req.Type == consts.Working {

		log.Debug("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
		needRes = false
		if s.Status != consts.Working {
			log.Warn("ProxyName [%s], is not working when it gets one new work conn", req.ProxyName)
		}

		s.GetNewCliConn(c)
	} else {
		info = fmt.Sprintf("ProxyName [%s], type [%d] unsupport", req.ProxyName, req.Type)
		log.Warn(info)
		return
	}
	succ = true
	return
}

func heart_beat(s *server.ProxyServer, c *conn.Conn) {
	isContinueRead := true
	f := func() {
		isContinueRead = false
		log.Error("ProxyName [%s], Server heartbeat tiemout", s.Name)
		s.Close()

	}

	timer := time.AfterFunc(time.Duration(server.HeartBeatTimeout)*time.Second, f)
	defer timer.Stop()

	res := &msg.ResponseMsg{
		Code: consts.HeartBeatCode,
	}
	resJson, _ := json.Marshal(res)

	for isContinueRead {
		content, err := c.ReadLine()
		if err != nil {
			if err == io.EOF {
				log.Warn("ProxyName [%s], client is dead !", s.Name)
				s.Close()
				break
			} else if c == nil || c.IsClosed() {
				log.Warn("ProxyName [%s], client connection is closed", s.Name)
				break
			}
			log.Error("ProxyName [%s], read error: %v", s.Name, err)
			continue
		}

		clientReq := &msg.RequestMsg{}
		if err := json.Unmarshal([]byte(content), &clientReq); err != nil {
			log.Warn("ProxyName [%s], Parse error: %v", s.Name, err)
			continue
		}

		if clientReq.Type != consts.HeartBeatType {
			continue
		}

		// log.Debug("ProxyName [%s], get heartbeat", s.Name)
		timer.Reset(time.Duration(server.HeartBeatTimeout) * time.Second)

		if err := c.Write(string(resJson) + "\n"); err != nil {
			log.Error("ProxyName [%s] send heartbeat response to client failed, err: %v", err)
			continue
		}

	}
}
