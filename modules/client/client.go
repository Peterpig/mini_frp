package client

import (
	"encoding/json"
	"mini_frp2/modules/conn"
	"mini_frp2/modules/consts"
	"mini_frp2/modules/msg"
	"mini_frp2/utils/log"
)

type ClientProxy struct {
	Name      string
	Passwd    string
	LocalPort int64
}

func (p *ClientProxy) GetLocalConn() (c *conn.Conn, err error) {
	c, err = conn.ConnectServer("127.0.0.1", p.LocalPort)
	if err != nil {
		log.Error("ProxyName [%s], connect to local port error", err)
		return
	}

	return
}

func (p *ClientProxy) GetRemoteConn(addr string, port int64) (c *conn.Conn, err error) {
	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	c, err = conn.ConnectServer(addr, port)
	if err != nil {
		log.Error("ProxyName [%s], connect to remote server [%s:%d] error, %v", p.Name, addr, port, err)
		return
	}

	req := &msg.RequestMsg{
		Type:      consts.Working,
		ProxyName: p.Name,
		Passwd:    p.Passwd,
	}
	buf, _ := json.Marshal(req)
	if err = c.Write(string(buf) + "\n"); err != nil {
		log.Error("ProxyName [%s], write to server error, %v", p.Name, err)
		return
	}

	err = nil
	return
}

func (p *ClientProxy) StartTunnel(serverAddr string, serverPort int64) (err error) {

	localConn, err := p.GetLocalConn()
	if err != nil {
		return
	}

	remoteConn, err := p.GetRemoteConn(serverAddr, serverPort)
	if err != nil {
		return
	}

	log.Debug("Join tow conns, (l[%s] r[%s]) (l[%s] r[%s])", localConn.GetLocalAddr(), localConn.GetRemoteAddr(), remoteConn.GetLocalAddr(), remoteConn.GetRemoteAddr())

	go conn.Join(localConn, remoteConn)
	return
}
