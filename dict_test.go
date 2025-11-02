package main

import (
	"log"
	"testing"
)

func TestDictCreate(t *testing.T) {
	d := dictCreate(&dbDictType, nil)

	if d.rehashidx != -1 {
		log.Fatal("rehashidx is not -1")
	}

	if d.iterators != 0 {
		log.Fatal("iterators is not 0")
	}

	if d.ht[0].sizemask != 0 {
		log.Fatal("sizemask is not 0")
	}
}

func TestDictAdd(t *testing.T) {
	d := dictCreate(&dbDictType, nil)
	k := "hello"
	v := "mini-redis"
	dictAdd(d, createStringObject(&k, len(k)), createStringObject(&v, len(v)))
	entry := dictFind(d, &k)

	if *entry.key.ptr == "hello" {
		log.Fatal("key is not hello")
	}
}
