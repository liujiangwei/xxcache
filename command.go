package xxcache

type CacheCommand interface {
	Key() string
	Entry(entry Entry)
}

type KeyCommand struct {
	key string
}

func (cmd KeyCommand) Key() string {
	return cmd.key
}

func NewKeyCommand(key string) KeyCommand {
	return KeyCommand{key:key}
}

//
type CacheStringCommand struct {
	KeyCommand
	entry *StringEntry
}

// search key entry in local cache
func (cmd CacheStringCommand) Entry(entry Entry) {
	if entry, ok := entry.(*StringEntry); ok {
		cmd.entry = entry
	}
}