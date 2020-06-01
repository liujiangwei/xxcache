package intset

import (
	"encoding/binary"
	"github.com/sirupsen/logrus"
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

	logrus.Infoln(encoding, length, len(encoded))
	switch encoding {
	case EncInt16:
		for i:=0; i< int(length); i++{
			start := 8 + int(encoding) * i
			end := 8 + int(encoding) * (i+1)

			num := binary.LittleEndian.Uint16([]byte(encoded[start:end]))
			set = append(set, int64(num))
		}
	case EncInt32:
	case EncInt64:
	}
	return set
}
