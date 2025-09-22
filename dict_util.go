// Package main 提供了访问Go语言字典内部结构的工具。
// 注意：此实现是一个教育性的简化版本，主要使用反射来模拟map的bucket结构。
package main

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"unsafe"
)

// bucketElement 表示存储在Redis字典bucket中的键值对。
type bucketElement struct {
	Key   interface{}
	Value interface{}
}

// getBucketCount 获取Redis对象字典的bucket数量
// 参数m: Redis对象字典
// 返回值: bucket数量
func getBucketCount(m map[string]*redisObject) (int, error) {
	_, err := validateMap(m)
	if err != nil {
		return 0, err
	}

	// 使用反射获取map的bucket数量
	mapValue := reflect.ValueOf(m)

	// 获取map指针
	mapPtr := mapValue.UnsafePointer()
	mapPtrUint := uintptr(mapPtr)

	// 访问B字段（偏移量9）- bucket数量的对数
	// 在Go 1.23.9版本中，B字段的偏移量是9字节
	BPtr := (*uint8)(unsafe.Pointer(mapPtrUint + 9))
	return 1 << *BPtr, nil // 2^B
}

// getElementsInBucketIndex 获取Redis对象字典中根据索引模拟分配的元素
// 注意：此函数使用简化的逻辑模拟bucket分配，而不是真正访问Go map底层的bucket结构
// 参数m: Redis对象字典
// 参数bucketIndex: bucket索引（0-7）
// 返回值: 根据索引模拟分配的键值对元素列表
func getElementsInBucketIndex(m map[string]*redisObject, bucketIndex int) ([]bucketElement, error) {
	mapValue, err := validateMap(m)
	if err != nil {
		return nil, err
	}

	// 检查空字典
	if mapValue.Len() == 0 {
		return []bucketElement{}, nil
	}

	// 验证bucket索引
	if bucketIndex < 0 {
		errMsg := fmt.Sprintf("bucket索引不能为负数: %d", bucketIndex)
		log.Println(errMsg)
		return nil, errors.New(errMsg)
	}

	// 使用反射来模拟功能，并根据索引过滤元素
	elements := make([]bucketElement, 0)
	mapIter := mapValue.MapRange()

	// 模拟bucket索引分配（简化实现）
	elementCount := 0
	for mapIter.Next() {
		// 模拟根据bucket索引过滤元素
		if elementCount%8 == bucketIndex%8 { // 简化的bucket分配逻辑
			elements = append(elements, bucketElement{
				Key:   mapIter.Key().Interface(),
				Value: mapIter.Value().Interface(),
			})
		}
		elementCount++
	}

	return elements, nil
}

// getAllElementsInBuckets 获取Redis对象字典中所有bucket的元素
// 注意：此函数使用简化的逻辑模拟bucket分配，而不是真正访问Go map底层的bucket结构
// 参数m: Redis对象字典
// 返回值: 所有bucket中的键值对元素列表
func getAllElementsInBuckets(m map[string]*redisObject) ([][]bucketElement, error) {
	mapValue, err := validateMap(m)
	if err != nil {
		return nil, err
	}

	// 检查空字典
	if mapValue.Len() == 0 {
		return [][]bucketElement{}, nil
	}

	// 收集所有元素并按bucket分组（简化实现）
	buckets := make([][]bucketElement, 8) // Redis字典默认有8个bucket
	for i := range buckets {
		buckets[i] = make([]bucketElement, 0)
	}

	mapIter := mapValue.MapRange()
	elementCount := 0

	for mapIter.Next() {
		// 根据元素计数分配到不同的bucket
		bucketIdx := elementCount % 8
		buckets[bucketIdx] = append(buckets[bucketIdx], bucketElement{
			Key:   mapIter.Key().Interface(),
			Value: mapIter.Value().Interface(),
		})
		elementCount++
	}

	return buckets, nil
}

// validateMap 验证输入的Redis对象字典是否有效
// 参数m: 待验证的Redis对象字典
// 返回值: 字典的反射值和错误信息
func validateMap(m map[string]*redisObject) (reflect.Value, error) {
	if m == nil {
		err := errors.New("字典不能为空")
		log.Println(err)
		return reflect.Value{}, err
	}

	mapValue := reflect.ValueOf(m)
	if mapValue.Kind() != reflect.Map {
		err := errors.New("输入参数不是有效的字典类型")
		log.Println(err)
		return reflect.Value{}, err
	}

	return mapValue, nil
}
