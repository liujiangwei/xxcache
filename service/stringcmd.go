package service

import (
	"errors"
)

//  no type check

// "", error
// "nil", error
// "", nil
// "a",nil
//WRONGTYPE Operation against a key holding the wrong kind of value

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

func (server *Server) Get(key string) (string, error){
	v , ok := server.DatabaseSelected.Get(key)
	if !ok{
		return Nil, ErrKeyNotExist
	}

	if v, ok := v.(string); ok{
		return v, nil
	}else{
		return "", ErrWrongType
	}
}

func (server *Server)Set(key string, value string){
	server.DatabaseSelected.Set(key, value)
	return
}

var Nil = "Nil"

var ErrKeyNotExist = errors.New("error key is not exists")