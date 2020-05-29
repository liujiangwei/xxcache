package xxcache

import (
	"context"
	"errors"
	"github.com/liujiangwei/xxcache/redis"
)

type Pool struct {
	capacity int // max connection nums
	conn     chan *redis.Conn
	size     int // current size
	addr     string
}

func initPool(capacity int, addr string) (*Pool, error) {
	pool := Pool{
		capacity: capacity,
		conn:     nil,
		addr:     addr,
	}

	pool.conn = make(chan *redis.Conn, pool.capacity)

	for i := 0; i < pool.capacity; i++ {
		if conn, err := redis.Connect(pool.addr); err != nil {
			return nil, err
		} else {
			pool.conn <- conn
		}
	}

	return &pool, nil
}

func (pool *Pool) ExecCommand(ctx context.Context, command Command) {
	select {
	case conn := <-pool.conn:
		defer func() { pool.conn <- conn }()

		msg, err := conn.Send(command.Serialize())
		if err != nil {
			command.WithError(err)
			return
		}

		command.ParseResponse(msg)
	case <-ctx.Done():
		command.WithError(errors.New("redis connection timeout"))
	}
}

type ExecHandler func(conn *redis.Conn) error

func (pool *Pool) Exec(ctx context.Context, handler ExecHandler) error {
	select {
	case conn := <-pool.conn:
		defer func() { pool.conn <- conn }()

		return handler(conn)
	case <-ctx.Done():
		return errors.New("redis connection timeout")
	}
}
