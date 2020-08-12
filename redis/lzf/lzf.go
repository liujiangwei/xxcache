package lzf

/*
#include <lzf.h>
*/
import "C"
import (
	"unsafe"
)

func Compress(str string) string {
	encoded := make([]byte, len(str))
	input := []byte(str)
	length := C.lzf_compress(unsafe.Pointer(&input[0]), C.uint(len(str)), unsafe.Pointer(&encoded[0]), C.uint(len(str)))
	if length == 0 {
		return str
	}

	return string(encoded[0:length])
}

func DeCompress(str string, length int) string {
	input := []byte(str)
	var decoded = make([]byte, length)
	dl := C.lzf_decompress(unsafe.Pointer(&input[0]), C.uint(len(input)), unsafe.Pointer(&decoded[0]), C.uint(length))
	if dl == 0 {
		return str
	}
	return string(decoded)
}
