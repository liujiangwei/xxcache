package xxcache

import (
	"context"
	"errors"
	"github.com/liujiangwei/xxcache/redis"
)

type Pool struct {
	capacity int // max connection nums
	conn     chan *redis.Conn
	addr     string
}

func (pool *Pool)Init() error{
	if pool.addr == ""{
		return errors.New("addr required")
	}

	pool.conn = make(chan *redis.Conn, pool.capacity)
	for i := 0; i < pool.capacity; i++ {
		if conn, err := redis.Connect(pool.addr); err != nil {
			return err
		} else {
			pool.conn <- conn
		}
	}

	return nil
}


func (pool *Pool) ExecCommand(ctx context.Context, command redis.Command) {
	select {
	case conn := <-pool.conn:
		defer func() { pool.conn <- conn }()
		msg, err := conn.SendAndWaitReply(command.Serialize())
		if err != nil {
			command.Error(err)
			return
		}

		switch msg.(type) {
		case redis.ErrorMessage:
			command.Error(errors.New(msg.String()))
		case redis.NilMessage:
			command.Error(errors.New(msg.String()))
		default:
			command.Parse(msg)
		}
	case <-ctx.Done():
		command.Error(errors.New("redis connection timeout"))
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
