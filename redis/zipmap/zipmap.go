package zipmap

import (
	"bytes"
	"encoding/binary"
	"errors"
)

///Memory layout of a zipmap, for the map "foo" => "bar", "hello" => "world":
//<zmlen><len>"foo"<len><free>"bar"<len>"hello"<len><free>"world"
//zmlen 1 byte , only when less than 254
//len 1 or 5 bytes, if less than 254 only 1 byte or the next 4 bytes
//free
const BigLen = 254

type zipMap struct {
	value []byte
}

type iterator struct {
	cursor    int // 当前游标
	zm *zipMap
}

func (iterator *iterator)Next() (str string, ok bool){
	return str, ok
}

const End = 255

func New() zipMap {
	return zipMap{value: []byte{0 ,End}}
}

func (zm *zipMap)Iterate()(iterator iterator) {
	return
}

func (zm *zipMap)Set(key string, value string) error{
	// 需要的长度
	if _, ok := zm.Lookup(key); !ok{
		buf := &bytes.Buffer{}
		// key length
		if len(key) >= BigLen{
			buf.WriteByte(BigLen)
			if err := binary.Write(buf, binary.LittleEndian, uint32(len(key))); err != nil{
				return err
			}
		}else{
			buf.WriteByte(uint8(len(key)))
		}

		//key
		if n, err := buf.WriteString(key); err != nil{
			return err
		}else if n != len(key){
			return errors.New("write string err for key")
		}

		// value length
		if len(value) >= BigLen{
			buf.WriteByte(BigLen)
			if err := binary.Write(buf, binary.LittleEndian, uint32(len(value))); err != nil{
				return err
			}
		}else{
			buf.WriteByte(uint8(len(value)))
		}
		//value
		if n, err := buf.WriteString(value); err != nil{
			return err
		}else if n != len(value){
			return errors.New("write string err for key")
		}

		zm.value = append(zm.value[: len(zm.value)-1], buf.Bytes()...)
		zm.value = append(zm.value, End)

		if zm.value[0] < BigLen {
			zm.value[0]++
		}
	}else{

	}

	return nil
}

func (zm *zipMap)Del(key string) {

}

func (zm *zipMap)Get(key string) (value string, ok bool){
	return value, ok
}

func (zm *zipMap)Len() int{
	return 0
}

func (zm *zipMap)BlobLen() int{
	return 0
}

func (zm *zipMap)Lookup(key string) (n int, ok bool){
	for i :=1; i< len(zm.value) && zm.value[i] != End; i++{
		// key length
		kl := int(zm.value[i])
		i++
		if kl >= BigLen{
			kl = int(binary.LittleEndian.Uint32(zm.value[i: i+4]))
			i += 4
		}

		// key
		k := string(zm.value[i:i+kl])
		if key == k{
			return i, true
		}
		i+=kl

		// value length
		vl := int(zm.value[i])
		i++
		if vl >= BigLen{
			vl = int(binary.LittleEndian.Uint32(zm.value[i: i+4]))
			i += 4
		}

		// value free length
		vfl := int(zm.value[i])
		i++
		if vfl >= BigLen{
			vfl = int(binary.LittleEndian.Uint32(zm.value[i: i+4]))
			i += 4
		}

		// value
		i+= vl
		// value free
		i+= vfl
	}

	return len(zm.value), false
}


func Len(key string, value string) int{
	l := len(key) + len(value) + 3
	if len(key) >= BigLen{
		l += 4
	}

	if len(value) >= BigLen{
		l += 4
	}

	return l
}

func Load(encoded string) (hash map[string]string){
	hash = make(map[string]string, encoded[0])
	for i:=1; i< len(encoded)-1; i++{
		kl := uint32(encoded[i])
		i++
		if kl >= BigLen{
			kl = binary.LittleEndian.Uint32([]byte(encoded[i : i+4]))
			i += 4
		}
		// key
		key := encoded[i : i+int(kl)]
		i+=int(kl)

		// value length
		vl := uint32(encoded[i])
		i++
		if vl >= BigLen{
			vl = binary.LittleEndian.Uint32([]byte(encoded[i : i+4]))
			i += 4
		}

		// value free length
		vfl := uint32(encoded[i])
		i++
		if vfl >= BigLen{
			vfl = binary.LittleEndian.Uint32([]byte(encoded[i : i+4]))
			i += 4
		}

		// value
		value := encoded[i: i+int(vl)]
		i+= int(vl)
		// value free
		i+= int(vfl)

		hash[key] = value
	}

	return hash
}