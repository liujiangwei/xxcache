package ziplist

import (
	"encoding/binary"
	"strconv"
)

// <zlbytes><zltail><zllen><entry>...<entry><zlend>
//51 0 0 0 48 0 0 0 20 0
//0 241 2 241
//2 242 2 242
//2 243 2 243
//2 244 2 244
//2 245 2 245
//2 246 2 246
//2 247 2 247
//2 248 2 248
//2 249 2 249
//2 250 2 250
//255
const BigPrevLen = 254
const ZipEnd = 255

const ZipStrMask = 0xc0
const ZIP_INT_MASK = 0x30
const ZipStr06b = 0 << 6
const ZipStr14b = 1 << 6
const ZipStr32b = 2 << 6
const ZipInt16b = 0xc0 | 0<<4
const ZipInt32b = 0xc0 | 1<<4
const ZipInt64b = 0xc0 | 2<<4
const ZipInt24b = 0xc0 | 3<<4
const ZipInt8b = 0xfe

type Iterator struct {
	val string
	pos int
}

func (iterator *Iterator) Length() uint16 {
	str := iterator.val[8:10]

	return binary.LittleEndian.Uint16([]byte(str))
}

func newIterator(str string) Iterator {
	iterator := Iterator{
		val: str,
		pos: 10,
	}

	return iterator
}

func (iterator *Iterator) Next() (str string, ok bool) {
	if iterator.val[iterator.pos] == ZipEnd {
		return "", false
	}

	var length, size uint32
	// prev size
	_, size = iterator.prevLength()
	iterator.pos += int(size)

	length, size, str = iterator.entry()
	iterator.pos += int(length + size)
	return str, true
}

func (iterator *Iterator) prevLength() (length, size uint32) {
	length = uint32(iterator.val[iterator.pos])
	size = 1

	if length >= BigPrevLen {
		size += 4
		length = binary.LittleEndian.Uint32([]byte(iterator.val[iterator.pos+1 : iterator.pos+5]))
	}

	return length, size
}

// length entry value length
// size entry length size
// str entry value
func (iterator *Iterator) entry() (size, length uint32, str string) {
	encoding := iterator.decode()
	if encoding < ZipStrMask {
		switch encoding {
		case ZipStr06b:
			size = 1
			length = uint32(iterator.val[iterator.pos] & 0x3f)
		case ZipStr14b:
			size = 2
			//(len) = (((ptr)[0] & 0x3f) << 8) | (ptr)[1];
			length = uint32((iterator.val[iterator.pos]&0x3f)<<8 | iterator.val[iterator.pos+1])
		case ZipStr32b:
			size = 5
			pos := iterator.pos
			length = uint32(iterator.val[pos+1]<<24 | iterator.val[pos+2]<<16 | iterator.val[pos+3]<<8 | iterator.val[pos+4])
		default:
			panic("failed to decode zip list item")
		}
	} else {
		size = 1
		switch encoding {
		case ZipInt8b:
			length = 1
		case ZipInt16b:
			length = 2
		case ZipInt24b:
			length = 3
		case ZipInt32b:
			length = 4
		case ZipInt64b:
			length = 5
		default:
			if encoding >= ZipIntImmMin && encoding <= ZipIntImmMax {
				val := (encoding & ZIP_INT_IMM_MASK)-1
				str = strconv.Itoa(int(val))
			}else{
				panic("failed to decode zip list item")
			}
		}
	}

	if length > 0{
		str  = iterator.val[iterator.pos + 1: iterator.pos + int(length)]
	}

	return length, size, str
}

const ZipIntImmMin = 0xf1
const ZipIntImmMax = 0xfd
const ZIP_INT_IMM_MASK = 0x0f

func (iterator *Iterator) decode() uint8 {
	encoding := iterator.val[iterator.pos]
	if encoding < ZipStrMask {
		encoding = encoding & ZipStrMask
	}

	return encoding
}

//  if ((encoding) < ZIP_STR_MASK) {                                           \
//        if ((encoding) == ZIP_STR_06B) {                                       \
//            (lensize) = 1;                                                     \
//            (len) = (ptr)[0] & 0x3f;                                           \
//        } else if ((encoding) == ZIP_STR_14B) {                                \
//            (lensize) = 2;                                                     \
//            (len) = (((ptr)[0] & 0x3f) << 8) | (ptr)[1];                       \
//        } else if ((encoding) == ZIP_STR_32B) {                                \
//            (lensize) = 5;                                                     \
//            (len) = ((ptr)[1] << 24) |                                         \
//                    ((ptr)[2] << 16) |                                         \
//                    ((ptr)[3] <<  8) |                                         \
//                    ((ptr)[4]);                                                \
//        } else {                                                               \
//            panic("Invalid string encoding 0x%02X", (encoding));               \
//        }                                                                      \
//    } else {                                                                   \
//        (lensize) = 1;                                                         \
//        (len) = zipIntSize(encoding);                                          \
//    }

func Load(encoded string) (list []string){
	iter := newIterator(encoded)

	for str, ok := iter.Next(); ok; str, ok = iter.Next(){
		list = append(list, str)
	}

	return list
}
