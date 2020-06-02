package lzf

import "testing"

func TestCompress(t *testing.T) {
	t.Log(Compress("aabbaabbaa"))
}

func TestDeCompress(t *testing.T) {
	str := "aabbaabbaa"

	encoded := Compress(str)

	t.Log(DeCompress(encoded, len(str)))
}