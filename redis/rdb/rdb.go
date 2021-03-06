package rdb

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/liujiangwei/xxcache/redis/lzf"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"strconv"
	"unsafe"
)

const OpCodeExpireTime = 253
const OpCodeExpireTimeMs = 252
const OpCodeFreq = 249
const OpCodeIdle = 248
const OpCodeEof = 255
const OpCodeSelectDB = 254
const OpCodeResizeDB = 251
const OpCodeAux = 250
const OpCodeModuleAux = 247

func LoadOpCode(reader *bufio.Reader) (uint, error) {
	p, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return uint(p), nil
}

const EncVal = 3 //RDB_ENCVAL
const BitLen6 = 0
const BitLen14 = 1
const BitLen32 = 0x80
const BitLen64 = 0x81

func LoadLen(reader *bufio.Reader) (length uint64, encoded bool, err error) {
	var buf = make([]byte, 2)
	if buf[0], err = reader.ReadByte(); err != nil {
		return
	}

	// compare first 2 bit
	switch (buf[0] & 0xC0) >> 6 {
	case EncVal:
		encoded = true
		length = uint64(buf[0] & 0x3F)
	case BitLen6:
		length = uint64(buf[0] & 0x3F)
	case BitLen14:
		buf[1], err = reader.ReadByte()
		if err == nil {
			//l := (buf[0] & 0x3F << 8) | buf[1]
			num := binary.BigEndian.Uint16([]byte{buf[0] & 0x3F, buf[1]})
			length = uint64(num)
		} else {
			panic("failed to load BitLen14")
		}
	default:
		// compare first byte
		switch buf[0] {
		case BitLen32:
			logrus.Infoln("BitLen32")
			buf = make([]byte, 4)
			if n, err := io.ReadFull(reader, buf); err == nil && n == 4 {
				var lengthUint32 uint32
				err = binary.Write(bytes.NewBuffer(buf), binary.LittleEndian, lengthUint32)
				length = uint64(lengthUint32)
			} else if n != 4 {
				err = errors.New("failed to read 4 byte for BitLen32")
			}
		case BitLen64:
			logrus.Infoln("BitLen64")

			buf = make([]byte, 8)
			if n, err := io.ReadFull(reader, buf); err == nil && n == 8 {
				err = binary.Write(bytes.NewBuffer(buf), binary.LittleEndian, length)
			} else if n != 8 {
				err = errors.New("failed to read 8 byte for BitLen64")
			}
		default:
			err = errors.New(fmt.Sprintf("Unknown length encoding %d in rdbLoadLen(), %d, %d", buf[0], BitLen32, BitLen64))
		}
	}

	return length, encoded, err
}

const EncInt8 = 0
const EncInt16 = 1
const EncInt32 = 2
const EncLzf = 3

func LoadString(reader *bufio.Reader) (str string, err error) {
	var length uint64
	var encoded bool
	if length, encoded, err = LoadLen(reader); err != nil {
		return str, err
	}

	if encoded {
		switch length {
		case EncInt8:
			buf, err := reader.ReadByte()
			if err != nil {
				return "", err
			}
			str = fmt.Sprintf("%d", int8(buf))
		case EncInt16:
			// 2 byte
			var buf = make([]byte, 2)
			if n, err := io.ReadFull(reader, buf); err != nil {
				return str, err
			} else if n != 2 {
				return str, errors.New("failed to read 2 bytes for EncInt16")
			}

			num := binary.LittleEndian.Uint16(buf)
			str = fmt.Sprintf("%d", num)
		case EncInt32:
			// 4 byte
			var buf = make([]byte, 4)
			if n, err := io.ReadFull(reader, buf); err != nil {
				log.Fatal(err)
			} else if n != 4 {
				return str, errors.New("failed to read 4 bytes for EncInt32")
			}
			num := binary.LittleEndian.Uint32(buf)
			//buf[0] | (buf[1] << 8) | (buf[2] << 16 | buf[3] << 24)
			str = fmt.Sprintf("%d", num)
		case EncLzf:
			var n uint64
			// 解压前长度
			n, _, err = LoadLen(reader)
			if err != nil {
				return
			}

			// 解压后长度
			length, _, err := LoadLen(reader)
			if err != nil {
				return str, err
			}

			var c = make([]byte, n)
			if n, err := io.ReadFull(reader, c); err != nil {
				return str, err
			} else if n != len(c) {
				return str, errors.New(fmt.Sprintf("failed to read [%d] bytes for LoadString EncLzf", len(c)))
			}

			str = lzf.DeCompress(string(c), int(length))
		default:
			err = errors.New("failed decode string")
		}

		logrus.Debugln("load string length[%d]", length)
		return str, err
	}

	buf := make([]byte, length)
	if _, err := reader.Read(buf); err != nil {
		return str, err
	}
	str = string(buf)

	return str, err
}

//R_Zero = 0.0;
//R_PosInf = 1.0/R_Zero;
//RNegInf = -1.0/R_Zero;
//RNan = R_Zero/R_Zero;
// todo
const RZero = float64(1.0)
const RPosInf = 1.0 / RZero
const RNegInf = -1.0 / RZero
const RNan = RZero / RZero

func LoadDouble(reader *bufio.Reader) (f float64, err error) {
	var length byte
	if length, err = reader.ReadByte(); err != nil {
		return f, err
	}
	switch uint(length) {
	case 255:
		return RNegInf, nil
	case 254:
		return RPosInf, nil
	case 253:
		return RNan, nil
	default:
		var buf = make([]byte, length)
		var n int
		if n, err = io.ReadFull(reader, buf); err != nil {
			return f, err
		} else if n != int(length) {
			return f, errors.New(fmt.Sprintf("failed to load [%d] bytes for LoadDouble", int(length)))
		}
		return strconv.ParseFloat(string(buf), 64)
	}
}

func LoadBinaryDouble(reader *bufio.Reader) (f float64, err error) {
	length := unsafe.Sizeof(f)
	var buf = make([]byte, length)
	var n int
	if n, err = io.ReadFull(reader, buf); err != nil {
		return f, err
	} else if n != int(length) {
		return f, errors.New(fmt.Sprintf("failed to load [%d] bytes for LoadBinaryDouble", int(length)))
	}

	err = binary.Write(bytes.NewBuffer(buf), binary.LittleEndian, f)
	return f, err
}

const (
	TypeString          = 0
	TypeList            = 1
	TypeSet             = 2
	TypeZSet2           = 5
	TypeZSet            = 3
	TypeHash            = 4
	TypeListQuickList   = 14
	TypeHashZipMap      = 9
	TypeListZipList     = 10
	TypeSetIntSet       = 11
	TypeZSetZipList     = 12
	TypeHashZipList     = 13
	TypeStreamListPacks = 15
	TypeModule          = 6
	TypeModule2         = 7
)

const (
	ModuleOpCodeEof          = 0 /* End of module value. */
	RDB_MODULE_OPCODE_SINT   = 1 /* Signed integer. */
	ModuleOpCodeUint         = 2 /* Unsigned integer. */
	RDB_MODULE_OPCODE_FLOAT  = 3 /* Float. */
	RDB_MODULE_OPCODE_DOUBLE = 4 /* Double. */
	RDB_MODULE_OPCODE_STRING = 5 /* String. */
)
