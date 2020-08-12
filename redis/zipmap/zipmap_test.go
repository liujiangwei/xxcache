package zipmap

import (
	"log"
	"testing"
)

func TestZipMap_Set(t *testing.T) {
	zm := New()
	log.Println(zm.Set("a", "v"))
	log.Println(zm)
}
