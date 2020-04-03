package dict

import (
	"github.com/liujiangwei/xxcache/list"
)

type dict struct {
	HashFunc func(string) int
	Size int
	Values []*list.List
}

type DNode struct {
	key string
	Value interface{}
	Next *DNode
}

const DefaultSize = 100

func New() *dict {
	dict := new(dict)
	
	dict.Size = DefaultSize
	dict.Values = make([]*list.List, dict.Size)

	dict.HashFunc = func(s string) int {
		return 1
	}

	return dict
}