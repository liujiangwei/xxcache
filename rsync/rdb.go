package rsync

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/liujiangwei/xxcache/lzf"
	"github.com/liujiangwei/xxcache/obj"
	"github.com/liujiangwei/xxcache/redis"
	"log"
)

func NewRdbClient(addr string) (*Client, error) {
	client := new(Client)
	con, err := redis.Connect(addr)
	if err != nil {
		return nil, err
	}
	client.Conn = con

	return client, nil
}

type Client struct {
	addr string
	*redis.Conn
}

// 0000 0
// 0001 1
// 0010 2
// ...
// 1000 8
// 1001 9
// 1010 a
// 1011 b
// 1100 c
// 1101 d
// 1110 e
// 1111 f
func (client *Client) LoadOpCode() (uint, error) {
	p, err := client.Reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return uint(p), nil
}

func (client *Client) LoadLen() (length uint64, encoded bool, err error) {
	var buf = make([]byte, 2)
	if buf[0], err = client.Reader.ReadByte(); err != nil {
		return
	}
	// compare first 2 bit
	var dt = (buf[0] & 0xC0) >> 6
	log.Println("LoadLen", dt)

	if dt == RDB_ENCVAL {
		encoded = true
		length = uint64(buf[0] & 0x3F)
	} else if dt == RDB_6BITLEN {
		length = uint64(buf[0] & 0x3F)
		log.Println("RDB_6BITLEN", buf[0], buf[1])
	} else if dt == RDB_14BITLEN {
		buf[1], err = client.Reader.ReadByte()
		if err != nil {
			return
		}
		length = uint64((buf[0]&0x3F)<<8 | buf[1])
	} else if buf[0] == RDB_32BITLEN {
		buf = make([]byte, 4)
		if _, err = client.Reader.Read(buf); err != nil {
			return
		}

		var lengthUint32 uint32
		// change big to little endian
		binary.Write(bytes.NewBuffer(buf), binary.LittleEndian, lengthUint32)
		length = uint64(32)
	} else if buf[0] == RDB_64BITLEN {
		buf = make([]byte, 8)
		if _, err = client.Reader.Read(buf); err != nil {
			return
		}
		binary.Write(bytes.NewBuffer(buf), binary.LittleEndian, length)
	} else {
		err = errors.New(fmt.Sprintf("Unknown length encoding %d in rdbLoadLen()", dt))
	}

	return length, encoded, err
}

const RDB_TYPE_STRING = 0
const RDB_TYPE_LIST = 1
const RDB_TYPE_SET = 2
const RDB_TYPE_ZSET_2 = 5
const RDB_TYPE_ZSET = 3
const RDB_TYPE_HASH = 4
const RDB_TYPE_LIST_QUICKLIST = 14
const RDB_TYPE_HASH_ZIPMAP = 9
const RDB_TYPE_LIST_ZIPLIST = 10
const RDB_TYPE_SET_INTSET = 11
const RDB_TYPE_ZSET_ZIPLIST = 12
const RDB_TYPE_HASH_ZIPLIST = 13
const RDB_TYPE_STREAM_LISTPACKS = 15
const RDB_TYPE_MODULE = 6
const RDB_TYPE_MODULE_2 = 7

func (client *Client) LoadObj(rdbType uint) (*obj.Obj, error) {
	switch rdbType {
	case RDB_TYPE_STRING:
		val, err := client.LoadString()
		if err != nil {
			return nil, err
		}

		return obj.NewStringObj(val), nil
	case RDB_TYPE_LIST:
		length, _, err := client.LoadLen()
		if err != nil {
			return nil, err
		}

		log.Println("RDB_TYPE_LIST", length)
		for ; length > 0; length-- {
			if val, err := client.LoadString(); err != nil {
				return nil, err
			} else {
				log.Println("RDB_TYPE_LIST", val.String())
			}
		}
	case RDB_TYPE_SET:
		length, _, err := client.LoadLen()
		if err != nil {
			return nil, err
		}

		for ; length > 0; length-- {
			if val, err := client.LoadString(); err != nil {
				return nil, err
			} else {
				log.Println("RDB_TYPE_SET", val.String())
			}
		}
	case RDB_TYPE_ZSET, RDB_TYPE_ZSET_2:
		length, _, err := client.LoadLen()
		if err != nil {
			return nil, err
		}

		for ; length > 0; length-- {
			if val, err := client.LoadString(); err != nil {
				return nil, err
			} else {
				log.Println("RDB_TYPE_SET", val.String())
			}
		}
	case RDB_TYPE_HASH:
	case RDB_TYPE_LIST_QUICKLIST:
	case RDB_TYPE_HASH_ZIPMAP, RDB_TYPE_LIST_ZIPLIST, RDB_TYPE_SET_INTSET, RDB_TYPE_ZSET_ZIPLIST, RDB_TYPE_HASH_ZIPLIST:
	case RDB_TYPE_STREAM_LISTPACKS:
	case RDB_TYPE_MODULE, RDB_TYPE_MODULE_2:
	}

	return nil, nil
}

func (client *Client) LoadString() (*obj.StringObj, error) {
	length, encoded, err := client.LoadLen()
	if err != nil {
		return nil, err
	}

	if encoded {
		switch length {
		case RDB_ENC_INT8:
			buf, err := client.Reader.ReadByte()
			if err != nil {
				return nil, err
			}
			return obj.NewStringFromInt8(int8(buf)), nil
		case RDB_ENC_INT16:
			// 2 byte
			var buf = make([]byte, 2)
			if _, err := client.Reader.Read(buf); err != nil {
				return nil, err
			}
			num := binary.LittleEndian.Uint16(buf)

			return obj.NewStringFromInt16(int16(num)), nil
		case RDB_ENC_INT32:
			// 4 byte
			var buf = make([]byte, 4)
			if _, err := client.Reader.Read(buf); err != nil {
				log.Fatal(err)
			}
			num := binary.LittleEndian.Uint32(buf)
			//buf[0] | (buf[1] << 8) | (buf[2] << 16 | buf[3] << 24)
			return obj.NewStringFromInt32(int32(num)), nil

		case RDB_ENC_LZF:
			cLength, _, err := client.LoadLen()
			if err != nil {
				return nil, err
			}

			length, _, err := client.LoadLen()
			if err != nil {
				return nil, err
			}

			log.Println("RDB_ENC_LZF", cLength, length)
			var c = make([]byte, cLength)
			if _, err := client.Reader.Read(c); err != nil {
				return nil, err
			}

			str := lzf.DeCompress(string(c))
			return obj.NewStringFromString(str), nil
		}
	}

	var buf = make([]byte, length)
	if _, err := client.Reader.Read(buf); err != nil {
		return nil, err
	}

	return obj.NewStringFromString(string(buf)), nil
}

//
//func LoadInteger(reader *bufio.Reader, enctype uint64, flags int64) (stringo, error){
//	var val stringo
//	switch enctype {
//	case RDB_ENC_INT8:
//		b, err := reader.ReadByte()
//		return stringo(b), err
//	case RDB_ENC_INT16:
//		var bs = make([]byte, 2)
//		if l, err := reader.Read(bs); err != nil || l != 2{
//			return stringo(bs), err
//		}
//		val =  stringo(bs[0] | (bs[1] << 8))
//	case RDB_ENC_INT32:
//		var bs = make([]byte, 4)
//
//		val = stringo(bs[0] | (bs[1]<< 8 | bs[2]<< 16 | bs[3] << 24))
//	default:
//		return "", errors.NewConn("error enctype")
//	}
//
//	var plain = flags & RDB_LOAD_PLAIN
//	var sds = flags& RDB_LOAD_SDS
//	var encode = flags&RDB_LOAD_ENC
//
//	if plain == RDB_LOAD_PLAIN || sds == RDB_LOAD_SDS{
//
//	}else if encode == RDB_LOAD_ENC{
//
//	}else{
//
//	}
//}

func loadLzfString(reader *bufio.Reader, length uint64) {

}
