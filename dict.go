package main

import (
	"math"
)

const dict_hash_function_seed = 5381
const DICT_OK = 0
const DICT_ERR = 1
const DICT_HT_INITIAL_SIZE = 4
const dict_can_resize = 1
const dict_force_resize_ratio = 5

/**
 * 字典键值对定义
 */
type dictEntry struct {
	key  *robj
	val  *robj
	next *dictEntry
}

/**
 * 字典哈希表定义
 */
type dictht struct {
	table    *[]*dictEntry
	size     uint64
	sizemask int
	used     uint64
}

/**
 * 字典核心数据结构定义
 */
type dict struct {
	dType     *dictType
	privdata  *interface{}
	ht        *[2]dictht
	rehashidx int64
	iterators int
}

/**
 * 字典类型特定函数定义
 */
type dictType struct {
	hashFunction  func(key string) int
	keyDup        func(privdata *interface{}, key *interface{}) *interface{}
	valDup        func(privdata *interface{}, obj *interface{}) *interface{}
	keyCompare    func(privdata *interface{}, key1 string, key2 string) bool
	keyDestructor func(privdata *interface{}, key *interface{})
	valDestructor func(privdata *interface{}, obj *interface{})
}

func dictCreate(typePtr *dictType, privDataPtr *interface{}) *dict {
	d := new(dict)
	d.ht = &[2]dictht{}
	_dictInit(d, privDataPtr, typePtr)
	return d
}

func _dictInit(d *dict, privDataPtr *interface{},
	typePtr *dictType) int {

	_dictReset(&(d.ht)[0])
	_dictReset(&(d.ht)[1])
	d.privdata = privDataPtr
	d.dType = typePtr
	d.rehashidx = -1
	d.iterators = 0

	return DICT_OK

}

func _dictReset(ht *dictht) {
	ht.table = nil
	ht.size = 0
	ht.sizemask = 0
	ht.used = 0
}

func _dictExpandIfNeeded(d *dict) int {
	if dictIsRehashing(d) {
		return DICT_OK
	}

	if d.ht[0].size == 0 {
		dictExpand(d, DICT_HT_INITIAL_SIZE)
	}

	if d.ht[0].used >= d.ht[0].size &&
		(dict_can_resize == 1 || d.ht[0].used/d.ht[0].size > dict_force_resize_ratio) {
		dictExpand(d, d.ht[0].size<<1)
	}
	return DICT_OK
}

func dictAdd(d *dict, k *robj, v *robj) int {
	entry := dictAddRaw(d, k)
	if entry == nil {
		return DICT_ERR
	}
	entry.val = v
	return DICT_OK
}

func dictAddRaw(d *dict, k *robj) *dictEntry {

	if dictIsRehashing(d) {
		_dictRehashStep(d)
	}

	index := _dictKeyIndex(d, k)
	//检查key是否存在
	if index == -1 {
		return nil

	}
	//根据渐进式哈希表确定table
	var ht *dictht
	if dictIsRehashing(d) {
		ht = &d.ht[1]
	} else {
		ht = &d.ht[0]
	}

	//将key设置到对应table的拉链上，并维护必要的信息

	entry := &dictEntry{key: k}
	entry.next = (*(ht.table))[index]
	(*(ht.table))[index] = entry
	ht.used++

	return entry
}

func dictDelete(ht *dict, key string) int {
	return dictGenericDelete(ht, key, 0)
}

// 删除字典中的key
func dictGenericDelete(d *dict, k string, nofree int) int {
	if d.ht[0].size == 0 {
		return DICT_ERR
	}

	if dictIsRehashing(d) {
		_dictRehashStep(d)
	}

	h := dictGenHashFunction(k, len(k))
	var preDe *dictEntry

	for i := 0; i < 2; i++ {
		idx := h & d.ht[i].sizemask
		he := (*(d.ht[i].table))[idx]

		for he != nil {
			if (*he.key.ptr).(string) == k {
				if preDe != nil {
					preDe.next = he.next
				} else {
					(*(d.ht[0].table))[idx] = he.next
				}
				d.ht[i].used--
				if nofree != 0 {
					//help gc
					he = nil
				}
				return DICT_OK
			}
			preDe = he
			he = he.next
		}

		if !dictIsRehashing(d) {
			break
		}
	}

	return DICT_ERR
}

func dictReplace_new(d map[string]*robj, key *robj, val *robj) bool {
	k := (*key.ptr).(string)
	if _, e := d[k]; e {
		d[k] = val
		return false
	} else {
		d[k] = val
		return true
	}

}

func dictReplace(d *dict, key *robj, val *robj) bool {
	if dictAdd(d, key, val) == DICT_OK {
		return true
	}

	entry := dictFind(d, (*key.ptr).(string))
	if entry == nil {
		return false
	}
	entry.val = val
	return true

}

func dictFind(d *dict, key string) *dictEntry {

	if d.ht[0].used+d.ht[1].used == 0 {
		return nil
	}

	if dictIsRehashing(d) {
		_dictRehashStep(d)
	}

	h := dictGenHashFunction(key, len(key))
	for i := 0; i < 2; i++ {
		idx := h & d.ht[i].sizemask
		he := (*d.ht[0].table)[idx]
		for he != nil {
			if (*he.key.ptr).(string) == key {
				return he
			}
			he = he.next
		}
		if !dictIsRehashing(d) {
			break
		}
	}

	return nil

}

func dictIsRehashing(d *dict) bool {
	return d.rehashidx != -1
}

func _dictKeyIndex(d *dict, key *robj) int {

	var idx int
	if _dictExpandIfNeeded(d) == DICT_ERR {
		return -1
	}

	//计算索引位置
	h := d.dType.hashFunction((*key.ptr).(string))

	//基于索引定位key
	for i := 0; i < 2; i++ {
		idx = h & d.ht[i].sizemask
		he := (*(d.ht[i].table))[idx]

		for he != nil {
			if d.dType.keyCompare(nil, (*key.ptr).(string), (*he.key.ptr).(string)) {
				return -1
			}
			he = he.next
		}

		// 如果正在rehash，则检查ht[1]
		if !dictIsRehashing(d) {
			break
		}
	}

	return idx
}

func _dictRehashStep(d *dict) {
	if d.iterators == 0 {
		dictRehash(d, 1)
	}
}

// 渐进式哈希
func dictRehash(d *dict, n int) int {
	//最大容错次数
	empty_visits := n * 10

	if !dictIsRehashing(d) {
		return 0
	}
	//循环n次的渐进式重试，在最大限制内完成
	for n > 0 && d.ht[0].used != 0 {
		n--
		var de *dictEntry
		var nextde *dictEntry
		for (*(d.ht[0].table))[d.rehashidx] == nil {
			d.rehashidx++
			empty_visits--

			if empty_visits == 0 {
				return 1
			}
		}
		de = (*(d.ht[0].table))[d.rehashidx]

		for de != nil {
			nextde = de.next
			h := dictGenHashFunction((*de.key.ptr).(string), len((*de.key.ptr).(string))) & d.ht[1].sizemask
			de.next = (*(d.ht[1].table))[h]
			(*(d.ht[1].table))[h] = de

			d.ht[0].used--
			d.ht[1].used++

			de = nextde
		}
		(*(d.ht[0].table))[d.rehashidx] = nil
		d.rehashidx++
	}
	//原子交换判断
	if d.ht[0].used == 0 {
		d.ht[0].table = nil
		d.ht[0] = d.ht[1]
		_dictReset(&(d.ht[1]))
		d.rehashidx = -1
	}

	return 1

}

// 字典扩容
func dictExpand(d *dict, size uint64) int {
	var n dictht
	//获取实际空间
	realSize := _dictNextPower(size)
	//健壮性校验
	if dictIsRehashing(d) || d.ht[0].used > size {
		return DICT_ERR
	}

	if realSize == d.ht[0].size {
		return DICT_ERR
	}

	n.size = realSize
	n.sizemask = int(realSize - 1)
	n.table = &[]*dictEntry{}
	table := make([]*dictEntry, realSize)
	n.table = &table

	n.used = 0

	if d.ht[0].table == nil {
		d.ht[0] = n
		return DICT_OK
	}
	d.ht[1] = n
	d.rehashidx = 0

	return DICT_OK

}

func _dictNextPower(size uint64) uint64 {
	i := DICT_HT_INITIAL_SIZE

	if size >= math.MaxInt64 {
		return math.MaxInt64 + 1
	}

	for true {
		if uint64(i) >= size {
			break
		}
		i = i << 1
	}
	return uint64(i)
}
