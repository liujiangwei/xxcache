package xxcache

import (
	"context"
	"errors"
	"github.com/liujiangwei/xxcache/redis"
)

type Pool struct {
	capacity int // max connection nums
	conn chan *redis.Connection
	size int // current size
	addr string
}

func initPool(capacity int, addr string) (*Pool, error){
	pool := Pool{
		capacity: capacity,
		conn:     nil,
		addr:     addr,
	}

	pool.conn = make(chan *redis.Connection, pool.capacity)

	for i :=0; i< pool.capacity; i++{
		if conn, err := redis.Connect(pool.addr); err != nil{
			return nil, err
		}else{
			pool.conn <- conn
		}
	}

	return  &pool, nil
}

func (pool *Pool) Exec(ctx context.Context, command Command) error {
	select {
	case conn := <-pool.conn:
		defer func() {pool.conn <- conn}()

		if err := conn.Send(command.Serialize()); err != nil{
			return err
		}

		msg, err :=  conn.Recv()
		if err != nil {
			return err
		}

		return command.ParseResponse(msg)
	case <- ctx.Done():
		return errors.New("redis connection timeout")
	}
}