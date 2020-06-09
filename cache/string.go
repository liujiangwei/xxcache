package cache

type StringEntry struct {
	Val string
}

func (entry *StringEntry)Get() string {
	return entry.Val
}

func (entry *StringEntry)Set(val string)  {
	entry.Val = val
}