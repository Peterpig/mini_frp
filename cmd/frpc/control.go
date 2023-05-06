package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mini_frp2/modules/client"
	"mini_frp2/modules/conn"
	"mini_frp2/modules/consts"
	"mini_frp2/modules/msg"
	"mini_frp2/utils/log"
	"sync"
	"time"
)

var connection *conn.Conn = nil
var HeartBeatTimer *time.Timer = nil

func ControlProcess(clientProxy *client.ClientProxy, wait *sync.WaitGroup) {
	defer wait.Done()
	c, err := loginServer(clientProxy)
	if err != nil {
		log.Error("ProxyName [%s], connect to server [%s:%d] failed", clientProxy.Name, client.ServerAddr, client.ServerPort)
		return
	}

	connection = c
	defer connection.Close()

	for {
		content, err := connection.ReadLine()
		if err == io.EOF || connection == nil || connection.IsClosed() {
			log.Debug("ProxyName [%s], server close this control conn", clientProxy.Name)
			var sleepTime time.Duration = 1

			for {
				log.Debug("ProxyName [%s], try to reconnect to server [%s:%d]", clientProxy.Name, client.ServerAddr, client.ServerPort)
				c, err := loginServer(clientProxy)
				if err == nil {
					connection.Close()
					connection = c
					break
				}

				if sleepTime < 60 {
					sleepTime = sleepTime * 2
				}
				time.Sleep(sleepTime * time.Second)
			}
		} else if err != nil {
			log.Warn("ProxyName [%s], read from server error, %v", clientProxy.Name, err)
			continue
		}

		res := &msg.ResponseMsg{}
		if err := json.Unmarshal([]byte(content), &res); err != nil {
			log.Error("ProxyName [%s], parse err, %v", clientProxy.Name, err)
			continue
		}

		if res.Code == consts.HeartBeatCode {
			if HeartBeatTimer != nil {
				log.Debug("ProxyName [%s] rcv hertbeat resp sucess!", clientProxy.Name)
				HeartBeatTimer.Reset(time.Duration(client.HeartBeatTimeout) * time.Second)

			} else {
				log.Error("HeartBeatTimer is nil")
			}

			continue
		}

		clientProxy.StartTunnel(client.ServerAddr, client.ServerPort)
	}
}

func loginServer(clientProxy *client.ClientProxy) (c *conn.Conn, err error) {
	c, err = conn.ConnectServer(client.ServerAddr, client.ServerPort)
	if err != nil {
		log.Error("ProxyName [%s], connect to server [%s:%d] error, %v", clientProxy.Name, client.ServerAddr, client.ServerPort, err)
		return
	}

	req := &msg.RequestMsg{
		Type:      consts.ClientConn,
		ProxyName: clientProxy.Name,
		Passwd:    clientProxy.Passwd,
	}
	buf, _ := json.Marshal(req)
	err = c.Write(string(buf) + "\n")
	if err != nil {
		log.Error("ProxyName [%s], write to server error, %v", clientProxy.Name, err)
		return
	}

	res, err := c.ReadLine()
	if err != nil {
		log.Error("ProxyName [%s], read from server error, %v", clientProxy.Name, err)
		return
	}

	log.Debug("ProxyName [%s], read [%s]", clientProxy.Name, res)

	clientResp := &msg.ResponseMsg{}
	if err = json.Unmarshal([]byte(res), &clientResp); err != nil {
		log.Error("ProxyName [%s], format server response error, %v", clientProxy, err)
		return
	}

	if clientResp.Code != 0 {
		log.Error("ProxyName [%s], start proxy error, %v", clientProxy.Name, clientResp.Msg)
		return c, fmt.Errorf("%s", clientResp.Msg)
	}

	go startHeartBeat(clientProxy, c)
	log.Debug("ProxyName [%s], connect to server [%s:%d]", clientProxy.Name, client.ServerAddr, client.ServerPort)
	return
}

func startHeartBeat(clientProxy *client.ClientProxy, c *conn.Conn) {
	f := func() {
		log.Error("ProxyName [%s] Client HeartBeat timeout !", clientProxy.Name)
		if c != nil {
			c.Close()
		}
	}

	HeartBeatTimer = time.AfterFunc(time.Duration(client.HeartBeatTimeout)*time.Second, f)
	defer HeartBeatTimer.Stop()

	req := &msg.RequestMsg{
		Type:      consts.HeartBeatType,
		ProxyName: "",
		Passwd:    "",
	}

	reqJson, _ := json.Marshal(req)
	log.Debug("ProxyName [%s] start heartbeat", clientProxy.Name)

	for {
		time.Sleep(time.Duration(client.HeartBeatInterval) * time.Second)
		if c == nil || c.IsClosed() {
			break
		}

		if err := c.Write(string(reqJson) + "\n"); err != nil {
			log.Error("ProxyName [%s] send heartbeat to server error: %v", clientProxy.Name, err)
			continue
		}
		log.Debug("ProxyName [%s] send heartbeat success! ", clientProxy.Name)
		// timer.Reset(time.Duration(client.HeartBeatTimeout) * time.Second)
	}
	log.Debug("ProxyName [%s] heartbeat exit", clientProxy.Name)
}
