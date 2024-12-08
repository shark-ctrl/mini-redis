package main

import (
	"log"
	"testing"
)

func TestCreateZskipList(t *testing.T) {
	zsl := zslCreate()

	if zsl == nil {
		t.Error("zslCreate failed")
	}

	if zsl.header == nil {
		t.Error("跳表头节点为空")

	}

	if zsl.length > 0 {
		t.Error("跳表长度初始化失败")
	}

	if zsl.level != 1 {
		t.Error("跳表索引初始化失败")
	}

	if zsl.tail != nil {
		t.Error("跳表尾节点不为空")
	}

	if zsl.header.backward != nil {
		t.Error("头节点的后继节点不为空")
	}

}

func TestZslInsert(t *testing.T) {
	zsl := zslCreate()

	num := interface{}(1)
	obj := createObject(REDIS_ENCODING_INT, &num)
	zslInsert(zsl, 1.0, obj)

	num2 := interface{}(2)
	obj2 := createObject(REDIS_ENCODING_INT, &num2)
	zslInsert(zsl, 2.0, obj2)

	x := zsl.header
	for x != nil {
		log.Print("node:", x.obj, " span:", x.level[0].span)
		x = x.level[0].forward
	}

	log.Println("tail:", zsl.tail.obj)
	log.Println("length:", zsl.length)

	num1_5 := interface{}(1.5)
	obj1_5 := createObject(REDIS_ENCODING_INT, &num1_5)
	zslInsert(zsl, 1.5, obj1_5)

	printZskipListNode(zsl)

	rank := zslGetRank(zsl, 2.0, obj2)
	log.Println("rank:", rank)

	zslDelete(zsl, 1.5, obj1_5)
	log.Println("after delete print ****************")
	printZskipListNode(zsl)

}

func printZskipListNode(zsl *zskiplist) {
	log.Println("*******************")
	x := zsl.header
	for i := 0; i < zsl.level; i++ {
		x = zsl.header
		for x != nil {
			log.Print("node:", x.obj, " span:", x.level[i].span)
			x = x.level[i].forward
		}

		log.Println("*********** level", i, " end ***********")
	}
}
