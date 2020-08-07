package cache

type ListEntry struct {
	val []string
}

func (entry *ListEntry)AppendTail(val string){
	entry.val = append(entry.val, val)
}

func (entry *ListEntry)AppendHead(val string){
	entry.val = append([]string{val}, entry.val...)
}