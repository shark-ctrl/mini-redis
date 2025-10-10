package main

import "math"

const dict_hash_function_seed = 5381
const DICT_OK = 0
const DICT_ERR = 1
const DICT_HT_INITIAL_SIZE = 4

/**
 * 字典键值对定义
 */
type dictEntry struct {
	key  *string
	val  *interface{}
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
	hashFunction  func(key *string) int
	keyDup        func(privdata *interface{}, key *interface{}) *interface{}
	valDup        func(privdata *interface{}, obj *interface{}) *interface{}
	keyCompare    func(privdata *interface{}, key1 *string, key2 *string) bool
	keyDestructor func(privdata *interface{}, key *interface{})
	valDestructor func(privdata *interface{}, obj *interface{})
}

func dictCreate(typePtr *dictType, privDataPtr *interface{}) dict {
	d := &dict{}
	_dictInit(d, privDataPtr, typePtr)
	return *d
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

// 渐进式哈希
func dictRehash(d *dict, n int) int {
	//最大容错次数
	empty_visits := n * 10

	if dictIsRehashing(d) {
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
			h := dictGenHashFunction(de.key, len(*(de.key))&d.ht[1].sizemask)
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

func dictReplace(d map[string]*robj, key *robj, val *robj) bool {
	k := (*key.ptr).(string)
	if _, e := d[k]; e {
		d[k] = val
		return false
	} else {
		d[k] = val
		return true
	}
}

func dictAdd(d *dict, k *string, v *interface{}) int {
	entry := dictAddRaw(d, k)
	if entry == nil {
		return DICT_ERR
	}
	entry.val = v
	return DICT_OK
}

func dictAddRaw(d *dict, k *string) *dictEntry {
	//todo 判断是否需要渐进式哈希

	index := _dictKeyIndex(d, k)
	//检查key是否存在
	if index == -1 {
		return nil

	}
	//根据渐进式哈希表确定table
	var ht dictht
	if dictIsRehashing(d) {
		ht = d.ht[1]
	} else {
		ht = d.ht[0]
	}

	//将key设置到对应table的拉链上，并维护必要的信息

	entry := &dictEntry{key: k}
	entry.next = (*(ht.table))[index]
	ht.used++

	return entry
}

func dictIsRehashing(d *dict) bool {
	return d.rehashidx != -1
}

func _dictKeyIndex(d *dict, key *string) int {

	var idx int
	//todo 判断是否需要扩容

	//计算索引位置

	h := d.dType.hashFunction(key)

	//基于索引定位key
	for i := 0; i < 2; i++ {
		idx := h & d.ht[i].sizemask
		he := (*(d.ht[i].table))[idx]

		for he != nil {
			if d.dType.keyCompare(nil, key, key) {
				return -1
			}
			he = he.next
		}

		// 如果正在rehash，则检查ht[1]
		if dictIsRehashing(d) {
			break
		}
	}

	return idx
}
