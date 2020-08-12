package cache

import (
	"sync"
	"testing"
)

func TestDatabase_Incr(t *testing.T) {
	db := Cache{}

	wg := sync.WaitGroup{}
	wg.Add(200)
	for i := 0; i < 200; i++ {
		for i := 0; i < 100000; i++ {
			if _, err := db.IncrBy("a", 1); err != nil {
				t.Fatal(err)
			}
		}
		wg.Done()
	}
	wg.Wait()
	t.Log(db.Get("a"))
}
