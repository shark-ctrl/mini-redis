package main

import (
	"log"
)

func listTypePush(subject *robj, value *robj, where int) {

	if subject.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
	} else if subject.encoding == REDIS_ENCODING_LINKEDLIST {
		lobj := (*subject.ptr).(list)
		if where == REDIS_HEAD {
			listAddNodeHead(&lobj, value.ptr)
		} else {
			listAddNodeTail(&lobj, value.ptr)
		}
	} else {
		log.Panic("Unknown list encoding")
	}
}

func listTypeLength(subject *robj) int64 {

	if subject.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
		return 0
	} else if subject.encoding == REDIS_ENCODING_LINKEDLIST {
		lobj := (*subject.ptr).(list)
		return lobj.len
	} else {
		log.Panic("Unknown list encoding")
	}
	return -1
}

func listTypePop(subject *robj, where int) *robj {
	var value *robj
	if subject.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
	} else if subject.encoding == REDIS_ENCODING_LINKEDLIST {
		lobj := (*subject.ptr).(list)
		var ln *listNode
		if where == REDIS_HEAD {
			ln = lobj.head
		} else {
			ln = lobj.tail
		}
		v := (*ln.value).(robj)
		value = &v
		listDelNode(&lobj, ln)
	} else {
		log.Panic("Unknown list encoding")
	}
	return value
}
