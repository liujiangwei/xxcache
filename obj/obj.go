package obj

type objT uint

const (
	String objT = iota
	Hash
	Set
	ZSet
	List
)

type Obj struct {
	t   objT
	val interface{}
}

func NewStringObj(o *StringObj) *Obj {
	return &Obj{
		t:   String,
		val: o,
	}
}
