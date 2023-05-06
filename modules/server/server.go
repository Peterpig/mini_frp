package server

import (
	"container/list"
	"fmt"
	"mini_frp2/modules/conn"
	"mini_frp2/modules/consts"
	"mini_frp2/utils/log"
	"sync"
	"time"
)

type ProxyServer struct {
	Name     string
	Passwd   string
	BindAddr string
	BindPort int64

	Status int64

	linster *conn.Linster

	// 连接frp server的conn
	cliConnChan    chan *conn.Conn
	cliConnChanInt chan int64

	// 连接代理的conn
	userConnList *list.List
	mux          sync.Mutex
}

func (p *ProxyServer) Lock() {
	p.mux.Lock()
}

func (p *ProxyServer) UnLock() {
	p.mux.Unlock()
}

func (p *ProxyServer) Close() {
	p.Lock()
	defer p.UnLock()
	p.Status = consts.Idle
	p.linster.Close()
	close(p.cliConnChan)
	p.userConnList = list.New()
}

func (p *ProxyServer) Init() {
	p.Status = consts.Idle
	p.cliConnChan = make(chan *conn.Conn)
	p.userConnList = list.New()
	p.cliConnChanInt = make(chan int64)
}

func (p *ProxyServer) GetNewCliConn(c *conn.Conn) {
	p.cliConnChan <- c
}

func (p *ProxyServer) WaitUserConn() (closeFlag bool) {
	closeFlag = false

	_, ok := <-p.cliConnChanInt
	if !ok {
		closeFlag = true
	}
	return
}

func (p *ProxyServer) Start() (err error) {
	p.Init()
	p.linster, err = conn.Linsten(p.BindAddr, p.BindPort)
	if err != nil {
		return
	}

	p.Status = consts.Working
	log.Debug("ProxyName [%s], start working", p.Name)

	// 开始接受请求
	go func() {
		if p.BindPort != 7000 {
			fmt.Println("x")
		}
		c, err := p.linster.GetConn()
		if err != nil {
			log.Info("ProxyName [%s], linster is closed", p.Name)
			return
		}

		log.Debug("ProxyName [%s], get one user conn [%s]", p.Name, c.GetRemoteAddr())

		p.Lock()
		if p.Status != consts.Working {
			log.Debug("ProxyName [%s] is not working, new user conn close", p.Name)
			c.Close()
			p.UnLock()
			return
		}

		log.Debug("1111111")
		p.userConnList.PushBack(c)
		p.cliConnChanInt <- 1
		p.UnLock()
		log.Debug("33333333333333")

		time.AfterFunc(time.Duration(UserConnTimeout)*time.Second, func() {
			p.Lock()
			defer p.UnLock()
			element := p.userConnList.Front()
			if element == nil {
				return
			}

			userConn := element.Value.(*conn.Conn)
			if userConn == c {
				log.Warn("ProxyName [%s] user conn [%s] timeout", p.Name, c.GetRemoteAddr())
			}
		})

	}()

	// join连接
	go func() {
		for {
			cliConn, ok := <-p.cliConnChan
			if !ok {
				return
			}

			p.Lock()
			element := p.userConnList.Front()
			var userConn *conn.Conn
			if element != nil {
				userConn = element.Value.(*conn.Conn)
				p.userConnList.Remove(element)
			} else {
				cliConn.Close()
				p.UnLock()
				continue
			}
			p.UnLock()

			log.Debug("Join to conns, (l[%s] r[%s]) (l[%s] r[%s])", cliConn.GetLocalAddr(), cliConn.GetRemoteAddr(), userConn.GetLocalAddr(), userConn.GetRemoteAddr())

			go conn.Join(cliConn, userConn)
		}
	}()
	return
}
