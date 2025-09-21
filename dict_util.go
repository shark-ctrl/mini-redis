// Package main 提供了访问Go语言字典内部结构的工具。
//
// 注意：此实现是一个教育性的简化版本，主要使用反射来模拟map的bucket结构。
// 真实的实现需要使用unsafe包来访问Go运行时的hmap和bmap结构。
package main

import (
	"fmt"
	"reflect"
)

// DictUtil 是一个用于检查Go语言map内部结构的工具类。
type DictUtil struct{}

// NewDictUtil 创建一个新的DictUtil实例。
func NewDictUtil() *DictUtil {
	return &DictUtil{}
}

// BucketElement 表示存储在map bucket中的键值对。
type BucketElement struct {
	Key   interface{}
	Value interface{}
}

// GetBucketCount 返回map中的bucket数量。
// 注意：此实现是一个简化的版本，返回默认的bucket数量8。
// 在真实实现中，这会访问Go运行时的hmap结构来获取实际的bucket数量。
func (du *DictUtil) GetBucketCount(m interface{}) (int, error) {
	// 检查输入参数
	if m == nil {
		return 0, fmt.Errorf("map不能为空")
	}

	// 使用反射检查map
	mapValue := reflect.ValueOf(m)
	if mapValue.Kind() != reflect.Map {
		return 0, fmt.Errorf("输入的参数不是map类型")
	}

	// 在简化实现中，我们假设Go map默认有8个bucket
	// 在真实实现中，bucket数量是2的B次方（2^B）
	return 8, nil
}

// InspectMapBucket 检查map中指定索引的bucket。
// 注意：此实现是一个简化的版本，使用反射来模拟bucket结构。
// 在真实实现中，这会访问Go运行时的hmap结构来获取bucket信息。
func (du *DictUtil) InspectMapBucket(m interface{}, bucketIndex int) ([]BucketElement, error) {
	// 检查输入参数
	if m == nil {
		return nil, fmt.Errorf("map不能为空")
	}

	// 使用反射检查map
	mapValue := reflect.ValueOf(m)
	if mapValue.Kind() != reflect.Map {
		return nil, fmt.Errorf("输入的参数不是map类型")
	}

	// 检查空map
	if mapValue.Len() == 0 {
		return []BucketElement{}, nil
	}

	// 验证bucket索引
	if bucketIndex < 0 {
		return nil, fmt.Errorf("bucket索引不能为负数: %d", bucketIndex)
	}

	// 使用反射来模拟功能，并根据索引过滤元素
	elements := make([]BucketElement, 0)
	mapIter := mapValue.MapRange()

	// 模拟bucket索引分配（简化实现）
	elementCount := 0
	for mapIter.Next() {
		// 模拟根据bucket索引过滤元素
		if elementCount%8 == bucketIndex%8 { // 简化的bucket分配逻辑
			elements = append(elements, BucketElement{
				Key:   mapIter.Key().Interface(),
				Value: mapIter.Value().Interface(),
			})
		}
		elementCount++
	}

	return elements, nil
}

// InspectAllMapBuckets 检查map中的所有buckets。
// 注意：此实现是一个简化的版本，使用反射来模拟bucket结构。
func (du *DictUtil) InspectAllMapBuckets(m interface{}) ([][]BucketElement, error) {
	// 检查输入参数
	if m == nil {
		return nil, fmt.Errorf("map不能为空")
	}

	// 使用反射检查map
	mapValue := reflect.ValueOf(m)
	if mapValue.Kind() != reflect.Map {
		return nil, fmt.Errorf("输入的参数不是map类型")
	}

	// 检查空map
	if mapValue.Len() == 0 {
		return [][]BucketElement{}, nil
	}

	// 收集所有元素并按bucket分组（简化实现）
	buckets := make([][]BucketElement, 8) // Go map默认有8个bucket
	for i := range buckets {
		buckets[i] = make([]BucketElement, 0)
	}

	mapIter := mapValue.MapRange()
	elementCount := 0

	for mapIter.Next() {
		// 根据元素计数分配到不同的bucket
		bucketIdx := elementCount % 8
		buckets[bucketIdx] = append(buckets[bucketIdx], BucketElement{
			Key:   mapIter.Key().Interface(),
			Value: mapIter.Value().Interface(),
		})
		elementCount++
	}

	return buckets, nil
}
