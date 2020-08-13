package cache

//func TestHandleMessage(t *testing.T) {
//	cache := new(Cache)
//	for _, cmd := range testCommandList() {
//		if err := HandleMessage(cache, cmd.Serialize()); err != nil {
//			t.Fatal(cmd.Serialize(), err)
//		}
//	}
//}
//
//func testCommandList() []redis.BaseCommand {
//	return []redis.BaseCommand{
//		redis.NewBaseCommand("set", "a", "aa"),
//		redis.NewBaseCommand("set", "a", "aa"),
//		redis.NewBaseCommand("set", "a", "aa"),
//	}
//}
