package main

import "fmt"

func zslCreate() *zskiplist {

	zsl := new(zskiplist)

	zsl.level = 1
	zsl.length = 0
	zsl.header = zslCreateNode(ZSKIPLIST_MAXLEVEL, 0, nil)

	zsl.header.level = make([]zskiplistLevel, ZSKIPLIST_MAXLEVEL)
	for j := 0; j < ZSKIPLIST_MAXLEVEL; j++ {
		zsl.header.level[j].forward = nil
		zsl.header.level[j].span = 0
	}

	zsl.header.backward = nil
	zsl.tail = nil
	return zsl
}

func zslCreateNode(level int64, score float64, obj *robj) *zskiplistNode {
	zn := new(zskiplistNode)
	zn.level = make([]zskiplistLevel, level)
	zn.score = score
	zn.obj = obj
	return zn
}

/*
*
1. 从header开始遍历各层索引，找到各层小于x的最大节点并存到update中
2. 初始化节点
3. 基于update节点维护各层索引关系
4. 将高层索引进行span++
5. 维护当前节点的backward
6. 维护长度信息
*/
func zslInsert(zsl *zskiplist, score float64, obj *robj) *zskiplistNode {
	update := make([]*zskiplistNode, ZSKIPLIST_MAXLEVEL)
	rank := make([]int64, ZSKIPLIST_MAXLEVEL)
	x := zsl.header

	var j int64
	for j = zsl.level - 1; j >= 0; j-- {
		if x.level[j].forward != nil &&
			(x.level[j].forward.score < score || (x.level[j].forward.score == score && x.level[j].forward.obj.String() < obj.String())) {
			x = x.level[j].forward
			rank[j] += x.level[j].span
		}
		update[j] = x
	}

	//level:=
	//x=zslCreateNode()

	return x
}

func (o *robj) String() string {
	return fmt.Sprintf("%v", *o.ptr)
}
