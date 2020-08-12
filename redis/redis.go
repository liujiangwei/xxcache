package redis

import (
	"context"
	"strconv"
	"time"
)

func ConvertToMessage(args ...string) Message {
	msg := ArrayMessage{}
	for _, arg := range args {
		msg = append(msg, NewBulkStringMessage(arg))
	}
	return msg
}

type Redis struct {
	connPool Pool
}

type Option struct {
	Addr         string
	MaxRetry     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Timeout      time.Duration
	PoolSize     int
	Database     int
	Password     string
}

func NewClient(option Option) *Redis {
	redis := new(Redis)

	redis.connPool.Init(option.PoolSize, option.Addr)

	if option.Password != "" {
		redis.connPool.AddConnOpenHandler(func(conn *Conn) error {
			cmd := StringCommand{
				BaseCommand: NewBaseCommand("auth", option.Password),
			}

			conn.ExecCommand(context.Background(), &cmd)

			return cmd.Err
		})
	}

	if option.Database > 0 {
		redis.connPool.AddConnOpenHandler(func(conn *Conn) error {
			cmd := StringCommand{
				BaseCommand: NewBaseCommand("select", strconv.Itoa(option.Database)),
			}

			conn.ExecCommand(context.Background(), &cmd)

			return cmd.Err
		})
	}

	return redis
}

func (redis *Redis) process(command Command) {
	redis.connPool.ExecCommand(context.Background(), command)
}

// 1.search in local databases
// 2.search in remote redis server
func (redis *Redis) Set(key, value string) (string, error) {
	args := []string{"set", key, value}

	rc := OKCommand{}
	rc.BaseCommand = NewBaseCommand(args...)

	redis.process(&rc)

	return rc.Result()
}

// 1.search in local databases
// 2.search in remote redis server
func (redis *Redis) Select(index int) (string, error) {
	rc := OKCommand{
		BaseCommand: NewBaseCommand("select", strconv.Itoa(index)),
	}

	redis.process(&rc)

	return rc.Result()
}

func (redis *Redis) SetNX(key, value string) (string, error) {
	rc := OKCommand{
		BaseCommand: NewBaseCommand("SetNX", key, value),
	}
	redis.process(&rc)
	return rc.Result()
}

func (redis *Redis) SetEX(key, value string, expires uint64) (string, error) {
	rc := OKCommand{
		BaseCommand: NewBaseCommand("SetNX", key, value, strconv.FormatUint(expires, 10)),
	}
	redis.process(&rc)
	return rc.Result()
}

func (redis *Redis) PSetEX(key, value string, expiresMs uint64) (string, error) {
	rc := OKCommand{
		BaseCommand: NewBaseCommand("SetNX", key, value, strconv.FormatUint(expiresMs, 10)),
	}
	redis.process(&rc)
	return rc.Result()
}

func (redis *Redis) Get(key string) (string, error) {
	rc := StringCommand{}
	rc.BaseCommand = NewBaseCommand("get", key)
	redis.process(&rc)
	return rc.Val, rc.Err
}

func (redis *Redis) GetSet(key, value string) (string, error) {
	rc := StringCommand{
		BaseCommand: NewBaseCommand("GetSet", key, value),
	}
	redis.process(&rc)
	return rc.Result()
}

func (redis *Redis) StrLen(key string) (int, error) {
	rc := IntCommand{
		BaseCommand: NewBaseCommand("StrLen", key),
	}
	redis.process(&rc)

	return rc.Val, rc.Err
}

func (redis *Redis) Append(key, str string) (int, error) {
	rc := IntCommand{
		BaseCommand: NewBaseCommand("Append", key, str),
	}
	redis.process(&rc)

	return rc.Val, rc.Err
}

func (redis *Redis) SetRange(key string, offset int, value string) (int, error) {
	rc := IntCommand{
		BaseCommand: NewBaseCommand("SetRange", key, strconv.Itoa(offset), value),
	}
	redis.process(&rc)

	return rc.Val, rc.Err
}

func (redis *Redis) GetRange(key string, start, end int) (string, error) {
	rc := StringCommand{
		BaseCommand: NewBaseCommand("GetRange", key, strconv.Itoa(start), strconv.Itoa(end)),
	}
	redis.process(&rc)
	return rc.Result()
}

func (redis *Redis) Incr(key string) (int, error) {
	rc := IntCommand{
		BaseCommand: NewBaseCommand("Incr", key),
	}
	redis.process(&rc)

	return rc.Val, rc.Err
}

func (redis *Redis) IncrBy(key string, increment int) (int, error) {
	rc := IntCommand{
		BaseCommand: NewBaseCommand("Incr", key, strconv.Itoa(increment)),
	}
	redis.process(&rc)

	return rc.Val, rc.Err
}

func (redis *Redis) IncrByFloat(key string, increment float64) (float64, error) {
	rc := FloatCommand{
		BaseCommand: NewBaseCommand("IncrByFloat", key, strconv.FormatFloat(increment, 'f', -1, 64)),
	}
	redis.process(&rc)

	return rc.Val, rc.Err
}

func (redis *Redis) Decr(key string) (int, error) {
	rc := IntCommand{
		BaseCommand: NewBaseCommand("Decr", key),
	}
	redis.process(&rc)

	return rc.Val, rc.Err
}

func (redis *Redis) DecrBy(key string, decrement int) (int, error) {
	rc := IntCommand{
		BaseCommand: NewBaseCommand("DecrBy", key, strconv.Itoa(decrement)),
	}
	redis.process(&rc)

	return rc.Val, rc.Err
}

func (redis *Redis) MSet(kv map[string]string) (string, error) {
	var args = []string{"MSet"}
	for k, v := range kv {
		args = append(args, k, v)
	}

	rc := OKCommand{
		BaseCommand: NewBaseCommand(args...),
	}

	redis.process(&rc)

	return rc.Val, rc.Err
}

func (redis *Redis) MSetNX(kv map[string]string) (int, error) {
	var args = []string{"MSetNX"}
	for k, v := range kv {
		args = append(args, k, v)
	}

	rc := IntCommand{
		BaseCommand: NewBaseCommand(args...),
	}

	redis.process(&rc)

	return rc.Val, rc.Err
}

// error(nil) or string
func (redis *Redis) MGet(keys ...string) ([]interface{}, error) {
	rc := InterfaceArrayCommand{
		BaseCommand: NewBaseCommand(append([]string{"MGet"}, keys...)...),
	}
	redis.process(&rc)

	return rc.Val, rc.Err
}
