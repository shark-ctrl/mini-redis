package main

type listNode struct {
	prev  *listNode
	next  *listNode
	value *interface{}
}

type list struct {
	head *listNode
	tail *listNode
	len  int64
}

func listCreate() *list {
	var l *list
	l = new(list)
	l.head = nil
	l.tail = nil
	l.len = 0
	return l
}

func listAddNodeHead(l *list, value *interface{}) *list {
	var node *listNode
	node = new(listNode)
	node.value = value

	if l.len == 0 {
		l.head = node
		l.tail = node
	} else {
		node.prev = nil
		node.next = l.head
		l.head.prev = node
		l.head = node
	}
	l.len++
	return l
}

func listAddNodeTail(l *list, value *interface{}) *list {
	var node *listNode
	node = new(listNode)
	node.value = value

	if l.len == 0 {
		l.head = node
		l.tail = node
	} else {
		node.prev = l.tail
		node.next = nil
		l.tail.next = node
		l.tail = node
	}
	l.len++
	return l

}

func listInsertNode(l *list, old_node *listNode, value *interface{}, after bool) *list {
	var node *listNode
	node = new(listNode)
	node.value = value

	if after {
		node.prev = old_node
		node.next = old_node.next

		if l.tail == old_node {
			l.tail = node
		}
	} else {
		node.next = old_node
		node.prev = old_node.prev

		if l.head == old_node {
			l.head = node

		}
	}

	if node.prev != nil {
		node.prev.next = node
	}

	if node.next != nil {
		node.next.prev = node
	}

	l.len++
	return l

}

func listDelNode(l *list, node *listNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		l.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev
	}

	node.prev = nil
	node.next = nil

	l.len--

}

func listIndex(l *list, index int64) *listNode {
	var n *listNode
	if index < 0 {
		index = (-index) - 1
		n = l.tail

		for index > 0 && n != nil {
			n = n.prev
			index--
		}
	} else {
		n = l.head
		for index > 0 && n != nil {
			n = n.next
			index--
		}
	}

	return n
}
