package internal

import (
	"errors"
	"net"
	"sync"
	"time"
)

// Pool conn pool.
type Pool struct {
	New  func() (net.Conn, error)
	mu   sync.RWMutex
	size uint

	// TODO: 连接闲置到一定时间, 先关掉
	IdleTimeout time.Duration
	connCh      chan net.Conn
	limit       uint
	connTotal   uint
}

type wrapConn struct {
	net.Conn
	pool *Pool
}

// NewFunc new conn func.
type NewFunc func() (net.Conn, error)

// NewPool new conn pool.
func NewPool(newFunc NewFunc, size, limit uint, idleTimeout time.Duration) (*Pool, error) {
	p := &Pool{
		New:         newFunc,
		size:        size,
		limit:       limit,
		IdleTimeout: idleTimeout,
	}

	p.connCh = make(chan net.Conn, p.size)
	for i := 0; i < int(size); i++ {
		conn, err := p.New()
		if err != nil {
			return nil, err
		}
		p.connTotal++

		p.connCh <- conn
	}
	return p, nil
}

// Acquire get conn.
func (p *Pool) Acquire() (net.Conn, error) {
	select {
	case conn := <-p.connCh:
		if conn == nil {
			return nil, errors.New("close")
		}
		return newWrapConn(p, conn), nil
	default:
		p.mu.RLock()
		if p.connTotal >= p.limit {
			return nil, errors.New("conn total max")
		}
		p.mu.RUnlock()

		conn, err := p.New()
		if err != nil {
			return nil, err
		}
		p.mu.Lock()
		p.connTotal++
		p.mu.Unlock()

		return newWrapConn(p, conn), nil
	}
}

func (p *Pool) put(conn net.Conn) error {
	if conn == nil {
		return errors.New("conn closed")
	}

	select {
	case p.connCh <- conn:
		return nil
	default:
		p.mu.Lock()
		p.connTotal--
		p.mu.Unlock()
		return conn.Close()
	}
}

// Close close all conn.
func (p *Pool) Close() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	close(p.connCh)

	for conn := range p.connCh {
		conn.Close()
	}
}

func newWrapConn(p *Pool, conn net.Conn) net.Conn {
	wc := &wrapConn{pool: p}
	wc.Conn = conn
	return wc
}

func (wc *wrapConn) Close() error {
	return wc.pool.put(wc.Conn)
}
