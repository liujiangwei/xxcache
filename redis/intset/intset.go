package intset

import (
	"encoding/binary"
)

//#define INTSET_ENC_INT16 (sizeof(int16_t))
//#define INTSET_ENC_INT32 (sizeof(int32_t))
//#define INTSET_ENC_INT64 (sizeof(int64_t))

const(
	EncInt16 = 2
	EncInt32 = 4
	EncInt64 = 8
)

func Load(encoded string) (set []int64) {
	encoding := binary.LittleEndian.Uint32([]byte(encoded[0:4]))
	length := binary.LittleEndian.Uint32([]byte(encoded[4:8]))

	switch encoding {
	case EncInt16:
		for i:=0; i< int(length); i++{
			start := 8 + 2 * i
			end := 8 + 2 * (i+1)

			num := binary.LittleEndian.Uint16([]byte(encoded[start:end]))
			set = append(set, int64(num))
		}
	case EncInt32:
		for i:=0; i< int(length); i++{
			start := 8 + 4 * i
			end := 8 + 4 * (i+1)

			num := binary.LittleEndian.Uint32([]byte(encoded[start:end]))
			set = append(set, int64(num))
		}
	case EncInt64:
		for i:=0; i< int(length); i++{
			start := 8 + 8 * i
			end := 8 + 8 * (i+1)

			num := binary.LittleEndian.Uint64([]byte(encoded[start:end]))
			set = append(set, int64(num))
		}
	}

	return set
}
