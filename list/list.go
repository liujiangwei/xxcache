package list

type List struct {
	Head   *Node
	Tail   *Node
	Length int
}

type Node struct {
	Prev  *Node
	Next  *Node
	Value interface{}
}

func New() *List {
	return &List{
		Head:   nil,
		Tail:   nil,
		Length: 0,
	}
}

func (list *List) Empty() {
	for list.Length > 0 {
		list.Head = list.Head.Next
		list.Length--
	}
}

func (list *List) InsertToHead(v interface{}) {
	n := &Node{
		Prev:  nil,
		Next:  nil,
		Value: v,
	}

	if list.Length == 0 {
		list.Head, list.Tail = n, n
	} else {
		n.Next = list.Head
		list.Head.Prev = n
		list.Head = n
	}
	list.Length++
}

func (list *List) InsertToTail(v interface{}) {
	n := &Node{
		Prev:  nil,
		Next:  nil,
		Value: v,
	}

	if list.Length == 0 {
		list.Head, list.Tail = n, n
	} else {
		n.Prev = list.Tail
		list.Tail = n
		list.Tail.Prev = n
	}

	list.Length++
}

func (list *List) Insert(cur *Node, v interface{}, position int) {
	if cur == nil {
		return
	}

	p := cur
	for position != 0 {
		if position > 0 {
			if p.Next == nil {
				break
			} else {
				p = p.Next
				position--
			}
		} else if p.Prev != nil {
			if p.Prev == nil {
				break
			} else {
				p = p.Prev
				position++
			}
		}
	}

	n := &Node{
		Prev:  p,
		Next:  nil,
		Value: v,
	}

	if p.Next == nil {
		p.Next, list.Tail = n, n
	} else {
		n.Next.Prev = n
		n.Next = p.Next
		p.Next = n
	}

	list.Length++
}

func (list *List) Delete(node *Node) {
	if node.Prev != nil {
		node.Prev.Next = node
	} else {
		list.Head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		list.Tail = node.Prev
	}

	list.Length--
}
