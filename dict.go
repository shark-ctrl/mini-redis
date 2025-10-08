package main

const dict_hash_function_seed = 5381

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
	table    **dictEntry
	size     uint64
	sizemask uint64
	used     uint64
}

/**
 * 字典核心数据结构定义
 */
type dict struct {
	t         *dictType
	privdata  *interface{}
	ht        [2]dictht
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
