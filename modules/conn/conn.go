package conn

import (
	"bufio"
	"fmt"
	"io"
	"mini_frp2/utils/log"
	"net"
	"sync"
)

type Conn struct {
	TcpConn   *net.TCPConn
	Reader    *bufio.Reader
	closeFlag bool
}

type Linster struct {
	addr      net.Addr
	l         *net.TCPListener
	conns     chan *Conn
	closeFlag bool
}

func (l *Linster) Addr() string {
	return l.addr.String()
}

func (l *Linster) Close() {
	if l.l != nil && !l.closeFlag {
		l.closeFlag = true
		l.l.Close()
		close(l.conns)
	}
}

func (l *Linster) GetConn() (conn *Conn, err error) {
	var ok bool
	conn, ok = <-l.conns
	if !ok {
		return nil, fmt.Errorf("channel closed")
	}
	return
}

func (c *Conn) Close() {
	if c.TcpConn != nil && !c.closeFlag {
		c.closeFlag = true
		c.TcpConn.Close()
	}
}

func (c *Conn) IsClosed() bool {
	return c.closeFlag
}

func (c *Conn) Write(content string) (err error) {
	_, err = c.TcpConn.Write([]byte(content))
	return
}

func (c *Conn) ReadLine() (buff string, err error) {
	buff, err = c.Reader.ReadString('\n')
	if err == io.EOF {
		c.closeFlag = true
	}
	return
}

func (c *Conn) GetRemoteAddr() string {
	return c.TcpConn.RemoteAddr().String()
}

func (c *Conn) GetLocalAddr() string {
	return c.TcpConn.LocalAddr().String()
}

func Join(c1 *Conn, c2 *Conn) {
	var wait sync.WaitGroup
	pip := func(to *Conn, from *Conn) {
		defer to.Close()
		defer from.Close()
		defer wait.Done()

		var err error

		_, err = io.Copy(to.TcpConn, from.TcpConn)
		if err != nil {
			log.Warn("Join conns err")
		}
	}

	wait.Add(2)
	go pip(c1, c2)
	go pip(c2, c1)
}

func Linsten(addr string, port int64) (l *Linster, err error) {
	fmt.Printf("sssssssss = %s:%d \n", addr, port)
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", addr, port))
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return
	}

	l = &Linster{
		addr:      tcpAddr,
		l:         listener,
		conns:     make(chan *Conn),
		closeFlag: false,
	}

	go func() {
		for {
			conn, err := l.l.AcceptTCP()
			if err != nil {
				if l.closeFlag {
					return
				}
				continue
			}
			c := &Conn{
				TcpConn:   conn,
				closeFlag: false,
			}
			c.Reader = bufio.NewReader(c.TcpConn)
			l.conns <- c
		}
	}()
	return
}

func ConnectServer(host string, port int64) (c *Conn, err error) {
	c = &Conn{}

	serverAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		return
	}

	conn, err := net.DialTCP("tcp", nil, serverAddr)

	if err != nil {
		return
	}

	c.TcpConn = conn
	c.Reader = bufio.NewReader(c.TcpConn)
	c.closeFlag = false
	return
}
