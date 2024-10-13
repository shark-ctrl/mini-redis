package main

import "log"

func hashTypeLookupWriteOrCreate(c *redisClient, key *robj) *robj {
	o := lookupKeyWrite(c.db, key)

	if o == nil {
		r := createHashObject()
		dbAdd(c.db, key, r)
		return r
	}

	r := (*o).(*robj)
	if r.robjType != REDIS_HASH {
		addReplyError(c, shared.wrongtypeerr)
	}
	return nil

}

func createHashObject() *robj {
	o := new(robj)

	o.robjType = REDIS_HASH
	o.encoding = REDIS_ENCODING_HT
	dict := new(map[string]*interface{})

	i := interface{}(dict)
	o.ptr = &i

	return o
}

func hashTypeTryObjectEncoding(subject *robj, o1 *robj, o2 *robj) {
	if subject.encoding == REDIS_HASH {
		if o1 != nil {
			o1 = tryObjectEncoding(o1)
		}

		if o2 != nil {
			o2 = tryObjectEncoding(o2)
		}
	}
}

func hashTypeSet(o *robj, field *robj, value *robj) int {

	if o.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
		return 0
	} else if o.encoding == REDIS_ENCODING_HT {
		m := (*o.ptr).(map[string]*robj)
		if dictReplace(m, field, value) {
			return 0
		}
		return 1
	} else {
		log.Panic("Unknown hash encoding")
		return -1
	}
}

func hashTypeExists(o *robj, field *robj) bool {
	dict := (*o.ptr).(map[string]*robj)
	key := (*field.ptr).(string)
	if _, e := dict[key]; e {
		return true
	}
	return false
}
