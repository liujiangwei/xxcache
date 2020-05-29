package dict

import (
	"github.com/cornelk/hashmap"
)

type Dict interface {
	Get(interface{}) (interface{}, bool)
	Set(interface{}, interface{})
	Del(interface{})
	Len() int
}

func Default() Dict {
	return hashmap.New(100)
}
