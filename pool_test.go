package internal

import (
	"net"
	"testing"
	"time"
)

const addr = "127.0.0.1:8099"

func init() {
	_, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
}

func newConn() (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func TestNew(t *testing.T) {
	_, err := NewPool(newConn, 10, 100, 10*time.Minute)
	if err != nil {
		t.Error(err)
	}
}

func TestAcquire(t *testing.T) {
	p, err := NewPool(newConn, 10, 100, 10*time.Minute)
	if err != nil {
		t.Error(err)
	}

	if _, err := p.Acquire(); err != nil {
		t.Error(err)
	}
}

func TestCloseConn(t *testing.T) {
	p, err := NewPool(newConn, 10, 100, 10*time.Minute)
	if err != nil {
		t.Error(err)
	}

	conn, err := p.Acquire()
	if err != nil {
		t.Error(err)
	}

	if err := conn.Close(); err != nil {
		t.Error(err)
	}
}

func TestClosePool(t *testing.T) {
	p, err := NewPool(newConn, 10, 100, 10*time.Minute)
	if err != nil {
		t.Error(err)
	}
	p.Close()
}

func TestExceedSize(t *testing.T) {
	p, err := NewPool(newConn, 1, 100, 10*time.Minute)
	if err != nil {
		t.Error(err)
	}

	conn1, err := p.Acquire()
	defer conn1.Close()
	if err != nil {
		t.Fatal(err)
	}
	conn2, err := p.Acquire()
	defer conn2.Close()
	if err != nil {
		t.Fatal(err)
	}

	if p.connTotal != 2 {
		t.Error("连接数不对")
	}

}
