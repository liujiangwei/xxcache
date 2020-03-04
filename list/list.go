package list

type List struct {
	head   *node
	tail   *node
	length int
}

type node struct {
	prev  *node
	next  *node
	value string
}

func (node *node) Next() *node{
	return node.next
}

func (node *node) Prev() *node{
	return node.prev
}

func (node *node) Value() string{
	return node.value
}

func New() *List{
	tail := &node{}

	head := &node{
		next: tail,
	}

	tail.prev = head

	return &List{
		head:   head,
		tail:   tail,
		length: 0,
	}
}

func (list *List) InsertPrev(position *node, value string) *node {
	list.length++
	
	node := &node{
		prev:  position.prev,
		next:  position,
		value: value,
	}

	position.prev.next = node

	position.prev = node

	return node
}

func (list *List) InsertNext(position *node, value string) *node {
	list.length++

	node := &node{
		prev:  position,
		next:  position.next,
		value: value,
	}

	position.next.prev = node
	
	position.next = node

	return node
}

func (list *List) Remove(node *node) {
	list.length--

	node.prev.next = node.next

	node.next.prev = node.prev
}

func (list *List) Head() *node {
	return list.head
}

func (list *List) IsHead(node *node) bool {
	return list.head == node
}

func (list *List)Tail() *node {
	return list.tail
}

func (list *List)IsTail(node *node) bool {
	return list.tail == node
}

func (list *List)Length() int{
	return list.length
}


