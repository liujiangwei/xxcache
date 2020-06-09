package database

type StringEntry struct {
	data string
}

func (entry *StringEntry)Get() string{
	return entry.data
}

func (entry *StringEntry) Set(val string) {
	entry.data = val
}
