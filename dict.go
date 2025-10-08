package main

const dict_hash_function_seed = 5381
const DICT_OK = 0
const DICT_ERR = 1

/**
 * 字典键值对定义
 */
type dictEntry struct {
	key  *interface{}
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
	i := interface{}(k)
	entry := &dictEntry{key: &i}
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
