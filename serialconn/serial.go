package serialconn

import (
	"errors"
	"io"
	"sync"
	"wj/serial"
)

type Pool struct {
	address string
	maxNum  int
	pool    chan io.ReadWriteCloser
	lock    *sync.Mutex
}

func New(host string, num int) (p *Pool, err error) {
	if host == "" {
		err = errors.New("The address of server is empty.")
		return
	}
	p = new(Pool)
	p.address = host
	p.SetNumber(1)
	err = p.Initialize()
	return
}

func (this *Pool) Initialize() (err error) {
	this.lock = new(sync.Mutex)
	this.pool = make(chan io.ReadWriteCloser, this.GetNumber())
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

func (this *Pool) Get() (conn io.ReadWriteCloser, err error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if len(this.pool) == 0 {
		conn, err = serialer.Open(this.address)
		return
	}
	conn = <-this.pool
	return
}

func (this *Pool) Put(conn io.ReadWriteCloser) (err error) {
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
