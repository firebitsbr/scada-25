package tcpconn

import (
	"errors"
	"net"
	"strings"
	"sync"
	"wj/sock"
)

type Pool struct {
	address string
	bak     string
	cur     string
	maxNum  int
	pool    chan *net.TCPConn
	lock    *sync.Mutex
}

func New(host string, num int) (p *Pool, err error) {
	if host == "" {
		err = errors.New("The address of server is empty.")
		return
	}
	p = new(Pool)
	ss := strings.Split(host, ";")
	p.address = ss[0]
	p.cur = p.address
	if len(ss) > 1 {
		p.bak = ss[1]
	} else {
		p.bak = p.address
	}
	p.SetNumber(num)
	err = p.Initialize()
	return
}

func (this *Pool) Initialize() (err error) {
	this.lock = new(sync.Mutex)
	this.pool = make(chan *net.TCPConn, this.GetNumber())
	return
}

func (this *Pool) SetNumber(num int) {
	if num < 1 {
		num = 1
	}
	this.maxNum = num
}

func (this *Pool) GetNumber() int {
	if this.maxNum < 1 {
		this.maxNum = 1
	}
	return this.maxNum
}

func (this *Pool) SetAddress(address string) {
	this.address = address
}

func (this *Pool) GetAddress() string {
	return this.address
}

func (this *Pool) Lock() *sync.Mutex {
	return this.lock
}

func (this *Pool) Len() int {
	return len(this.pool)
}

func (this *Pool) Get() (conn *net.TCPConn, err error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if len(this.pool) == 0 {
		conn, err = sock.CreateTcpConnect(this.cur)
		if this.cur == this.address {
			this.cur = this.bak
		} else {
			this.cur = this.address
		}
		return
	}
	conn = <-this.pool
	return
}

func (this *Pool) Put(conn *net.TCPConn) (err error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if len(this.pool) >= this.maxNum {
		err = errors.New("Pool is full")
		return
	}
	this.pool <- conn
	return
}

func (this *Pool) Clear() {
	this.lock.Lock()
	defer this.lock.Unlock()

	close(this.pool)
	for conn := range this.pool {
		conn.Close()
	}
	this.pool = nil
}
