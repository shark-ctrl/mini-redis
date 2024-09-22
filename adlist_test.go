package main

import (
	"testing"
)

func TestListCreate(t *testing.T) {
	l := listCreate()
	if l.head != nil {
		t.Error("the head node is not nil")
	}
	if l.tail != nil {
		t.Error("the tail node is not nil")
	}
	if l.len != 0 {
		t.Error("linked list length initialization error.")
	}

}

func TestListAddNodeHead(t *testing.T) {
	l := listCreate()

	str := "first node"
	v := interface{}(&str)

	listAddNodeHead(l, &v)

	if l.head != l.tail {
		t.Error("inconsistency in initial insertion of head and tail nodes.")
	}
	var headValue *string
	headValue = (*l.head.value).(*string)

	if *headValue != "first node" {
		t.Error("failed to initialize value during head node insertion")
	}

	if l.len != 1 {
		t.Error("length remains unchanged after inserting an element")
	}

	secondNode := "second node"
	secV := interface{}(&secondNode)
	listAddNodeHead(l, &secV)

	headValue = (*l.head.value).(*string)
	if *headValue != "second node" {
		t.Error("failed to insert the head node.")
	}

	if l.len != 2 {
		t.Error("length remains unchanged after inserting an element")
	}

}

func TestListAddNodeTail(t *testing.T) {
	l := listCreate()

	secondNode := "second node"
	secV := interface{}(&secondNode)

	listAddNodeHead(l, &secV)

	firstNode := "first node"
	firstV := interface{}(&firstNode)

	listAddNodeHead(l, &firstV)

	thirdNode := "third node"
	thirdV := interface{}(&thirdNode)
	listAddNodeTail(l, &thirdV)

	if l.len != 3 {
		t.Error("failed to listAddNodeTail the tail node.")
	}

	tailValue := (*l.tail.value).(*string)

	if *tailValue != "third node" {
		t.Error("failed to listAddNodeTail the third node ")
	}

	tailPrevValue := (*l.tail.prev.value).(*string)
	if *tailPrevValue != "second node" {
		t.Error("failed to listAddNodeTail the third node ")
	}

}

func TestListIndex(t *testing.T) {
	l := listCreate()

	secondNode := "second node"
	secV := interface{}(&secondNode)

	listAddNodeHead(l, &secV)

	firstNode := "first node"
	firstV := interface{}(&firstNode)

	listAddNodeHead(l, &firstV)

	thirdNode := "third node"
	thirdV := interface{}(&thirdNode)
	listAddNodeTail(l, &thirdV)

	var node *listNode
	node = listIndex(l, 0)
	if *(*node.value).(*string) != "first node" {
		t.Error("incorrect query in listIndex.")
	}

	node = listIndex(l, 1)
	if *(*node.value).(*string) != "second node" {
		t.Error("incorrect query in listIndex.")
	}

	node = listIndex(l, 2)
	if *(*node.value).(*string) != "third node" {
		t.Error("incorrect query in listIndex.")
	}

	node = listIndex(l, -1)
	if *(*node.value).(*string) != "third node" {
		t.Error("incorrect query in listIndex.")
	}

	node = listIndex(l, -2)
	if *(*node.value).(*string) != "second node" {
		t.Error("incorrect query in listIndex.")
	}
}

func TestListDelNode(t *testing.T) {
	l := listCreate()

	secondNode := "second node"
	secV := interface{}(&secondNode)

	listAddNodeHead(l, &secV)

	firstNode := "first node"
	firstV := interface{}(&firstNode)

	listAddNodeHead(l, &firstV)

	thirdNode := "third node"
	thirdV := interface{}(&thirdNode)
	listAddNodeTail(l, &thirdV)

	fourthNode := "fourth node"
	fourthV := interface{}(&fourthNode)
	listAddNodeTail(l, &fourthV)

	delNode := listIndex(l, 1)
	listDelNode(l, delNode)

	if l.len != 3 {
		t.Fatal("listDelNode operation failed.")
	}

	listDelNode(l, l.head)
	if *(*l.head.value).(*string) != "third node" {
		t.Fatal("list del head node operation failed.")
	}

	listDelNode(l, l.tail)

	if *(*l.head.value).(*string) != "third node" {
		t.Fatal("list del tail node operation failed.")
	}
}

func TestListInsertNode(t *testing.T) {
	l := listCreate()

	secondNode := "second node"
	secV := interface{}(&secondNode)

	listAddNodeHead(l, &secV)

	firstNode := "first node"
	firstV := interface{}(&firstNode)

	listAddNodeHead(l, &firstV)

	thirdNode := "third node"
	thirdV := interface{}(&thirdNode)
	listAddNodeTail(l, &thirdV)

	fourthNode := "fourth node"
	fourthV := interface{}(&fourthNode)
	listAddNodeTail(l, &fourthV)

	var node *listNode
	node = listIndex(l, 0)

	insertNode := "insert node"
	insertNodeV := interface{}(&insertNode)
	listInsertNode(l, node, &insertNodeV, true)

	if *(*l.head.next.value).(*string) != "insert node" {
		t.Fatal("listInsertNode operation failed.")
	}

}
