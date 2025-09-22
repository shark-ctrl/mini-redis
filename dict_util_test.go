package main

import (
	"fmt"
	"log"
	"testing"
)

// TestValidateMap 测试 validateMap 函数
func TestValidateMap(t *testing.T) {
	// 测试用例1: 正常的字典
	t.Run("ValidMap", func(t *testing.T) {
		validMap := make(map[string]*redisObject)
		mapValue, err := validateMap(validMap)
		if err != nil {
			t.Errorf("validateMap() with valid map returned an error: %v", err)
		}

		// 打印 mapValue 的信息
		log.Printf("ValidMap - mapValue.Kind(): %v\n", mapValue.Kind())
		log.Printf("ValidMap - mapValue.Type(): %v\n", mapValue.Type())
		log.Printf("ValidMap - mapValue.IsValid(): %v\n", mapValue.IsValid())
	})

	// 测试用例2: nil 字典
	t.Run("NilMap", func(t *testing.T) {
		var nilMap map[string]*redisObject
		_, err := validateMap(nilMap)
		if err == nil {
			t.Error("validateMap() with nil map should return an error, but got nil")
		}
	})

	// 测试用例3: 非字典类型 (需要构造一个非字典类型的变量来测试)
	// 注意：在Go语言中，由于类型安全，直接传递非字典类型给函数会导致编译错误。
	// 因此，我们无法直接测试这种情况，除非使用 interface{} 并在运行时检查。
	// 但我们假设函数签名已经限制了输入为 map[string]*robj，所以这种情况不会发生。
}

// TestGetBucketCount 测试 getBucketCount 函数
func TestGetBucketCount(t *testing.T) {
	// 测试用例1: 空字典
	t.Run("EmptyMap", func(t *testing.T) {
		emptyMap := make(map[string]*redisObject)
		count, err := getBucketCount(emptyMap)
		if err != nil {
			t.Errorf("getBucketCount() with empty map returned an error: %v", err)
		}

		// 空map通常至少有一个bucket
		if count <= 0 {
			t.Errorf("getBucketCount() returned %d for empty map, expected > 0", count)
		}

		// 验证返回的bucket数量是2的幂次方
		if (count & (count - 1)) != 0 {
			t.Errorf("getBucketCount() returned %d, expected a power of 2", count)
		}
	})

	// 测试用例2: 有元素的字典
	t.Run("NonEmptyMap", func(t *testing.T) {
		nonEmptyMap := make(map[string]*redisObject)
		// 添加一些元素
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%d", i)
			nonEmptyMap[key] = &redisObject{}
		}

		count, err := getBucketCount(nonEmptyMap)
		if err != nil {
			t.Errorf("getBucketCount() with non-empty map returned an error: %v", err)
		}

		// 检查返回的bucket数量是否合理（应该是2的幂次方）
		if count <= 0 || (count&(count-1)) != 0 {
			t.Errorf("getBucketCount() returned %d, expected a power of 2", count)
		}

		t.Logf("Non-empty map with 10 elements has %d buckets", count)
	})

	// 测试用例3: 更多元素的字典
	t.Run("LargeMap", func(t *testing.T) {
		largeMap := make(map[string]*redisObject)
		// 添加更多元素以触发扩容
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key%d", i)
			largeMap[key] = &redisObject{}
		}

		count, err := getBucketCount(largeMap)
		if err != nil {
			t.Errorf("getBucketCount() with large map returned an error: %v", err)
		}

		// 检查返回的bucket数量是否合理（应该是2的幂次方）
		if count <= 0 || (count&(count-1)) != 0 {
			t.Errorf("getBucketCount() returned %d, expected a power of 2", count)
		}

		t.Logf("Large map with 100 elements has %d buckets", count)
	})

	// 测试用例4: nil 字典
	t.Run("NilMap", func(t *testing.T) {
		var nilMap map[string]*redisObject
		_, err := getBucketCount(nilMap)
		if err == nil {
			t.Error("getBucketCount() with nil map should return an error, but got nil")
		}
	})
}
