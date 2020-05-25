package lzf

import "log"

//https://segmentfault.com/a/1190000011425787
func DeCompress(compressed string) string {
	cl := len(compressed)
	if cl == 0 {
		return ""
	}

	var raw = make([]byte, 0)
	ip := 0
	op := 0

	for ip < cl{
		ctrl := compressed[ip]
		ip++
		log.Println("ctrl", ctrl)

		if ctrl < (1<<5){
			ctrl++
			for i:=uint8(ctrl); i>0; i--{
				if i <= ctrl{
					raw = append(raw, compressed[ip])
					ip++
					op++
				}
			}
			log.Println("str", len(raw), ip, op)
		}else{
			length := ctrl >> 5
			ref := op - int((ctrl & 0x1f) << 8) - 1
			log.Println("length 1", length)
			if length == 7{
				length += compressed[ip]
				ip++
			}

			ref -= int(compressed[ip])
			ip++
			log.Println("length 2", length)
			if length > 9{
				length += 2
				raw = append(raw, raw[(op-int(ref)):]...)
				op += ref
			}else{
				for i := int(length); i >= -1; i--{
					raw = append(raw, raw[ref])
					ref++
				}
			}
		}
	}
	log.Println(compressed,string(raw))
	return string(raw)
}

func Compress(raw string)string  {
	length := len(raw)
	if length == 1{
		return raw
	}

	return raw
	var ip = 0
	val := (raw[ip] << 8) | raw[ip+1]

	for ip < length-2{
		val = val << 8 | raw[ip+2]
		slot
	}
}

const HSize = 1 << HLog
const HLog = 16
func IDX(n uint) uint{
	a :=  ((n >> (3 *8-HLog))-n ) & (HSize -1)

	return a
}

func Next() {
	
}

func First(b []byte) byte{
	a :=  b[0] << 8 | b[1]

	return a
}