package cache

type Option struct {
	DatabaseSize int
	Port int
	Address string
}

var defaultOptionDatabaseSize = 8

func newOption() *Option{
	option := Option{

	}

	return &option
}