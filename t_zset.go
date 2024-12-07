package main

import (
	"fmt"
	"math/rand"
)

func zslCreate() *zskiplist {
	var j int

	zsl := new(zskiplist)

	zsl.level = 1
	zsl.length = 0
	zsl.header = zslCreateNode(ZSKIPLIST_MAXLEVEL, 0, nil)

	for j = 0; j < ZSKIPLIST_MAXLEVEL; j++ {
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

func zslInsert(zsl *zskiplist, score float64, obj *robj) *zskiplistNode {
	update := make([]*zskiplistNode, ZSKIPLIST_MAXLEVEL)
	rank := make([]int64, ZSKIPLIST_MAXLEVEL)

	x := zsl.header
	var i int64

	for i = zsl.level - 1; i >= 0; i-- {

		if i == zsl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		if x.level[i].forward != nil &&
			(x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.obj.String() < obj.String())) {

			rank[i] += x.level[i].span
			x = x.level[i].forward

		}
		update[i] = x
	}

	level := rand.Int63n(ZSKIPLIST_MAXLEVEL)
	if level > zsl.level {
		for i := zsl.level; i < level; i++ {
			rank[i] = 0
			update[i] = zsl.header
			update[i].level[i].span = zsl.length
		}
		zsl.level = level
	}

	x = zslCreateNode(level, score, obj)

	for i = 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = rank[0] - rank[i] + 1
	}
	//将当前高度到最新高度的节点跨度自增 即基于zsl.length+1
	for i := level; i < zsl.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == zsl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}

	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		zsl.tail = x
	}

	zsl.length++
	return x
}

func zslGetRank(zsl *zskiplist, score float64, obj *robj) int64 {
	var rank int64

	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		if x.level[i].forward != nil &&
			(x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.obj.String() < obj.String())) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		if x.level[i].forward.obj.String() == obj.String() {
			return rank
		}
	}
	return 0
}

func zslDelete(zsl *zskiplist, score float64, obj *robj) int64 {
	update := make([]*zskiplistNode, ZSKIPLIST_MAXLEVEL)

	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		if x.level[i].forward != nil &&
			(x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.obj.String() < obj.String())) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	x = x.level[0].forward
	if x != nil && x.level[0].forward.obj.String() == obj.String() {
		zslDeleteNode(zsl, x, update)
		return 1
	}
	return 0
}

func zslDeleteNode(zsl *zskiplist, x *zskiplistNode, update []*zskiplistNode) {

	var i int64
	for i = 0; i < zsl.level; i++ {
		if update[i].level[i].forward == x {

			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span -= 1
		}
	}

	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		zsl.tail = x.backward
	}

	for zsl.level > 1 && zsl.header.level[zsl.level-1].forward == nil {
		zsl.level--
	}

	zsl.length--

}

func (o *robj) String() string {
	return fmt.Sprintf("%v", *o.ptr)
}
