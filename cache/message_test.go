package cache

//func TestHandleMessage(t *testing.T) {
//	client := new(Cache)
//	for _, cmd := range testCommandList() {
//		if err := HandleMessage(client, cmd.Serialize()); err != nil {
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
