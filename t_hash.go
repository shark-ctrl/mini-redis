package main

import "log"

func hashTypeLookupWriteOrCreate(c *redisClient, key *robj) *robj {
	//check if the dictionary exists.
	o := lookupKeyWrite(c.db, key)
	//if it is nil, create a hash object and add to redisDb.
	if o == nil {
		o = createHashObject()
		dbAdd(c.db, key, o)
		return o
	}
	/**
	if it exists but is not a hash object,
	reply the user of a type error.
	*/
	if o.robjType != REDIS_HASH {
		addReply(c, shared.wrongtypeerr)
		return nil
	}
	return o
}

func createHashObject() *robj {
	o := new(robj)

	o.robjType = REDIS_HASH
	o.encoding = REDIS_ENCODING_HT

	dict := make(map[string]*robj)
	i := interface{}(dict)
	o.ptr = &i

	return o
}

func hashTypeTryObjectEncoding(subject *robj, o1 **robj, o2 **robj) {
	/**
	determine if the subject encoding  is a dict .
	if it is, proceed with the logic processing
	*/
	if subject.encoding == REDIS_ENCODING_HT {
		//perform type conversion on the field.
		if o1 != nil {
			*o1 = tryObjectEncoding(*o1)
		}
		//perform type conversion on the field.
		if o2 != nil {
			*o2 = tryObjectEncoding(*o2)
		}
	}
}

func hashTypeSet(o *robj, field *robj, value *robj) int {

	if o.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
		return 0
	} else if o.encoding == REDIS_ENCODING_HT {
		m := (*o.ptr).(map[string]*robj)
		/**
		if it is an add operation, return true,
		and then the function returns 1 .
		conversely,return 0 to inform the external system
		that the current operation is an update

		*/
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
	/**
	because the robj pointer records the interface type,
	when storing, the field is forcefully cast to the interface type,
	and the same applies to the key
	*/
	dict := (*o.ptr).(map[string]*robj)
	key := (*field.ptr).(string)
	//if it exists, return true.
	if _, e := dict[key]; e {
		return true
	}
	return false
}

func hashTypeDelete(o *robj, field *robj) bool {
	var deleted bool
	if o.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
		return false
	} else if o.encoding == REDIS_ENCODING_HT {
		dict := (*o.ptr).(map[string]*robj)
		key := (*field.ptr).(string)
		_, ok := dict[key]
		if ok {
			delete(dict, key)
			deleted = true
		}
	} else {
		log.Panic("Unknown hash encoding")
	}
	return deleted
}
