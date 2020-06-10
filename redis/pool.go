package redis

import (
	"context"
	"errors"
	"sync"
)

const defaultPoolCapacity = 20

type Pool struct {
	capacity         int // max connection nums
	size             int // current size
	connC            chan *Conn
	addr             string
	connOpenHandler  []func(conn *Conn) error
	connCloseHandler []func(conn *Conn) error
	l                sync.Mutex
}

func (pool *Pool) AddConnOpenHandler(handler func(conn *Conn) error) {
	pool.connOpenHandler = append(pool.connOpenHandler, handler)
}

func (pool *Pool) Init(capacity int, addr string) {
	pool.capacity = capacity
	pool.addr = addr

	if pool.capacity <= 0 {
		pool.capacity = defaultPoolCapacity
	}

	pool.connC = make(chan *Conn, pool.capacity)
}

func (pool *Pool) Get(ctx context.Context) (*Conn, error) {
	select {
	case c := <-pool.connC:
		return c, nil
	default:
		// no conn available, try create new conn
		pool.l.Lock()
		if pool.size < pool.capacity {
			conn, err := pool.create()
			if err != nil {
				return nil, err
			}
			pool.size++
			pool.connC <- conn
		}
		pool.l.Unlock()
	}

	select {
	case <-ctx.Done():
		return nil, errors.New("no connection available")
	case c := <-pool.connC:
		return c, nil
	}
}

func (pool Pool) create() (conn *Conn, err error) {
	if conn, err = Connect(pool.addr); err != nil {
		return conn, err
	}

	for _, handler := range pool.connOpenHandler {
		if err := handler(conn); err != nil {
			return conn, err
		}
	}

	return conn, err
}

func (pool Pool) Put(conn *Conn) {
	if conn != nil {
		pool.connC <- conn
	}
}

//Close close all conn opened in the pool
func (pool Pool) Close() error {
	for conn := range pool.connC {
		for _, handler := range pool.connCloseHandler {
			if err := handler(conn); err != nil {
				return err
			}
		}

		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (pool *Pool) ExecCommand(ctx context.Context, command Command) {
	conn, err := pool.Get(ctx)
	defer pool.Put(conn)

	if err != nil {
		command.Error(err)
		return
	}

	conn.ExecCommand(ctx, command)
}

type ExecHandler func(conn *Conn) error

func (pool *Pool) Exec(ctx context.Context, handler ExecHandler) error {
	conn, err := pool.Get(ctx)
	defer pool.Put(conn)

	if err != nil {
		return err
	}

	return handler(conn)
}
