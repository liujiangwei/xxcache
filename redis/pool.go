package redis

import (
	"context"
	"errors"
)

type Pool struct {
	Capacity int // max connection nums
	conn     chan *Conn
	Addr     string
}

func (pool *Pool)Init() error{
	if pool.Addr == ""{
		return errors.New("addr is required")
	}

	pool.conn = make(chan *Conn, pool.Capacity)
	for i := 0; i < pool.Capacity; i++ {
		if conn, err := Connect(pool.Addr); err != nil {
			return err
		} else {
			pool.conn <- conn
		}
	}

	return nil
}


func (pool *Pool) ExecCommand(ctx context.Context, command Command) {
	select {
	case conn := <-pool.conn:
		defer func() { pool.conn <- conn }()
		msg, err := conn.SendAndWaitReply(command.Serialize())
		if err != nil {
			command.Error(err)
			return
		}

		switch msg.(type) {
		case ErrorMessage:
			command.Error(errors.New(msg.String()))
		case NilMessage:
			command.Error(errors.New(msg.String()))
		default:
			command.Parse(msg)
		}
	case <-ctx.Done():
		command.Error(errors.New("redis connection timeout"))
	}
}

type ExecHandler func(conn *Conn) error

func (pool *Pool) Exec(ctx context.Context, handler ExecHandler) error {
	select {
	case conn := <-pool.conn:
		defer func() { pool.conn <- conn }()

		return handler(conn)
	case <-ctx.Done():
		return errors.New("redis connection timeout")
	}
}
