package obj

import (
	"strconv"
)

type stringT uint

const (
	StringTString stringT = iota
	StringTInt
	StringTUInt
	StringTInt8
	StringTUInt8
	StringTInt16
	StringTUInt16
	StringTInt32
	StringTUInt32
	StringTInt64
	StringTUInt64
)

type StringObj struct {
	t   stringT
	val string
}

func (o StringObj) String() string {
	return o.val
}

func (o StringObj) Int() (int, error) {
	if num, err := strconv.Atoi(o.val); err != nil {
		return 0, err
	} else {
		return num, nil
	}
}

func (o StringObj) Val() interface{} {
	return o.val
}

func NewStringFromString(str string) *StringObj {
	return &StringObj{
		t:   StringTString,
		val: str,
	}
}

func NewStringFromInt8(val int8) *StringObj {
	return &StringObj{
		t:   StringTInt8,
		val: strconv.FormatInt(int64(val), 10),
	}
}

func NewStringFromInt16(val int16) *StringObj {
	return &StringObj{
		t:   StringTInt16,
		val: strconv.FormatInt(int64(val), 10),
	}
}

func NewStringFromUInt16(val uint16) *StringObj {
	return &StringObj{
		t:   StringTUInt16,
		val: strconv.FormatUint(uint64(val), 10),
	}
}

func NewStringFromInt32(val int32) *StringObj {
	return &StringObj{
		t:   StringTInt32,
		val: strconv.FormatInt(int64(val), 10),
	}
}
