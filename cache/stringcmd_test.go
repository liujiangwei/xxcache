package cache

import "testing"

func BenchmarkServer_Set(b *testing.B) {
	b.SetParallelism(4)

	server := &Cache{}
	server.Start()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			server.Set("a", "b")
		}
	})
}

func BenchmarkServer_Incr(b *testing.B) {
	b.SetParallelism(4)

	server := &Cache{}
	server.Start()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			server.Incr("int")
		}
	})
}

func BenchmarkServer_Get(b *testing.B) {
	b.SetParallelism(2)

	server := &Cache{}
	server.Start()

	server.Set("int", "1")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			server.Get("int")
		}
	})
}