package lzf

import "testing"

func TestCompress(t *testing.T) {
	var str = "abc"

	t.Log(Compress(str))
}

func TestDeCompress(t *testing.T) {

}
