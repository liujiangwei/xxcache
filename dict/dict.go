package dict

import (
	"github.com/cornelk/hashmap"
)

type Dict interface {
	Get(interface{}) (interface{}, bool)
	Set(interface{}, interface{})
}

func Default() Dict{
	return hashmap.New(100)
}