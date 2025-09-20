// Package main 提供了 MapInspector 的单元测试。
package main

import (
	"testing"
)

// TestMapInspector_InspectMapBucket 测试 InspectMapBucket 方法
func TestMapInspector_InspectMapBucket(t *testing.T) {
	// 创建一个新的 MapInspector 实例
	inspector := NewMapInspector()

	// 创建一个示例 map 用于测试
	sampleMap := map[string]int{
		"apple":  5,
		"banana": 3,
		"orange": 8,
		"grape":  12,
		"kiwi":   7,
	}

	// 测试 1: 检查特定 bucket 中的元素
	t.Log("=== 测试 1: 检查 bucket 0 中的元素 ===")
	elements, err := inspector.InspectMapBucket(sampleMap, 0)
	if err != nil {
		t.Errorf("错误: %v", err)
	} else {
		t.Logf("Bucket 0 中的元素: %+v", elements)
		t.Logf("元素数量: %d", len(elements))
	}

	// 测试 2: 负数 bucket 索引的错误处理
	t.Log("=== 测试 2: 负数 bucket 索引的错误处理 ===")
	_, err = inspector.InspectMapBucket(sampleMap, -1)
	if err != nil {
		t.Logf("负数 bucket 索引的预期错误: %v", err)
	} else {
		t.Error("期望出现错误，但没有错误")
	}

	// 测试 3: nil map 的错误处理
	t.Log("=== 测试 3: nil map 的错误处理 ===")
	_, err = inspector.InspectMapBucket(nil, 0)
	if err != nil {
		t.Logf("nil map 的预期错误: %v", err)
	} else {
		t.Error("期望出现错误，但没有错误")
	}
}

// TestMapInspector_InspectAllMapBuckets 测试 InspectAllMapBuckets 方法
func TestMapInspector_InspectAllMapBuckets(t *testing.T) {
	// 创建一个新的 MapInspector 实例
	inspector := NewMapInspector()

	// 创建一个示例 map 用于测试
	sampleMap := map[string]int{
		"apple":  5,
		"banana": 3,
		"orange": 8,
		"grape":  12,
		"kiwi":   7,
	}

	// 测试: 检查所有 buckets 中的元素
	t.Log("=== 测试: 检查所有 buckets 中的元素 ===")
	allBuckets, err := inspector.InspectAllMapBuckets(sampleMap)
	if err != nil {
		t.Errorf("错误: %v", err)
	} else {
		t.Logf("Buckets 数量: %d", len(allBuckets))
		for i, bucket := range allBuckets {
			t.Logf("Bucket %d: %+v", i, bucket)
		}
	}
}

// TestMapInspector_DifferentMapTypes 测试不同 map 类型的处理
func TestMapInspector_DifferentMapTypes(t *testing.T) {
	// 创建一个新的 MapInspector 实例
	inspector := NewMapInspector()

	// 测试: 使用不同的 map 类型
	t.Log("=== 测试: 使用不同的 map 类型 ===")
	intMap := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}

	elements, err := inspector.InspectMapBucket(intMap, 0)
	if err != nil {
		t.Errorf("错误: %v", err)
	} else {
		t.Logf("来自 int map 的元素: %+v", elements)
	}
}

// BenchmarkMapInspector_InspectMapBucket 基准测试 InspectMapBucket 方法
func BenchmarkMapInspector_InspectMapBucket(b *testing.B) {
	// 创建一个新的 MapInspector 实例
	inspector := NewMapInspector()

	// 创建一个示例 map 用于基准测试
	sampleMap := map[string]int{
		"apple":  5,
		"banana": 3,
		"orange": 8,
		"grape":  12,
		"kiwi":   7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.InspectMapBucket(sampleMap, 0)
	}
}

// BenchmarkMapInspector_InspectAllMapBuckets 基准测试 InspectAllMapBuckets 方法
func BenchmarkMapInspector_InspectAllMapBuckets(b *testing.B) {
	// 创建一个新的 MapInspector 实例
	inspector := NewMapInspector()

	// 创建一个示例 map 用于基准测试
	sampleMap := map[string]int{
		"apple":  5,
		"banana": 3,
		"orange": 8,
		"grape":  12,
		"kiwi":   7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.InspectAllMapBuckets(sampleMap)
	}
}