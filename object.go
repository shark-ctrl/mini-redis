package main

func createStringObject(ptr *string, len int) *robj {
	return createEmbeddedStringObject(ptr, len)
}

func createEmbeddedStringObject(ptr *string, len int) *robj {
	o := new(robj)
	o.robjType = REDIS_STRING
	o.encoding = REDIS_ENCODING_EMBSTR
	i := interface{}(*ptr)
	o.ptr = &i
	return o

}
