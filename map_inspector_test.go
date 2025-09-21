// Package main 提供了 MapInspector 的单元测试。
package main

import (
	"fmt"
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

// TestMapInspector_GetBucketCount 测试 GetBucketCount 方法
func TestMapInspector_GetBucketCount(t *testing.T) {
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

	// 测试: 获取 bucket 数量
	t.Log("=== 测试: 获取 bucket 数量 ===")
	count, err := inspector.GetBucketCount(sampleMap)
	if err != nil {
		t.Errorf("错误: %v", err)
	} else {
		t.Logf("Bucket 数量: %d", count)
		if count != 8 {
			t.Errorf("期望的 bucket 数量是 8，但得到了 %d", count)
		}
	}

	// 测试: nil map 的错误处理
	t.Log("=== 测试: nil map 的错误处理 ===")
	_, err = inspector.GetBucketCount(nil)
	if err != nil {
		t.Logf("nil map 的预期错误: %v", err)
	} else {
		t.Error("期望出现错误，但没有错误")
	}

	// 测试: 非 map 类型的错误处理
	t.Log("=== 测试: 非 map 类型的错误处理 ===")
	_, err = inspector.GetBucketCount("not a map")
	if err != nil {
		t.Logf("非 map 类型的预期错误: %v", err)
	} else {
		t.Error("期望出现错误，但没有错误")
	}
}

// TestMapInspector_BucketCountConsistency 测试默认情况下map的bucket数和实现的函数数是否一致
func TestMapInspector_BucketCountConsistency(t *testing.T) {
	// 创建一个新的 MapInspector 实例
	inspector := NewMapInspector()

	// 测试不同大小的map
	testCases := []struct {
		name string
		size int
	}{
		{"小map", 5},
		{"中等map", 50},
		{"大map", 500},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建指定大小的map
			testMap := make(map[string]int)
			for i := 0; i < tc.size; i++ {
				testMap[fmt.Sprintf("key%d", i)] = i
			}

			// 获取bucket数量
			bucketCount, err := inspector.GetBucketCount(testMap)
			if err != nil {
				t.Errorf("获取bucket数量时出错: %v", err)
				return
			}

			// 检查bucket数量是否为8（简化实现中的默认值）
			if bucketCount != 8 {
				t.Errorf("期望bucket数量为8，但得到了%d", bucketCount)
			}

			// 验证InspectAllMapBuckets返回的bucket数量是否一致
			allBuckets, err := inspector.InspectAllMapBuckets(testMap)
			if err != nil {
				t.Errorf("InspectAllMapBuckets时出错: %v", err)
				return
			}

			if len(allBuckets) != bucketCount {
				t.Errorf("InspectAllMapBuckets返回的bucket数量(%d)与GetBucketCount返回的数量(%d)不一致", len(allBuckets), bucketCount)
			}

			// 验证所有bucket的索引都在有效范围内
			totalElements := 0
			for i, bucket := range allBuckets {
				totalElements += len(bucket)
				_, err := inspector.InspectMapBucket(testMap, i)
				if err != nil {
					t.Errorf("InspectMapBucket索引%d时出错: %v", i, err)
				}
			}

			// 验证元素总数是否正确
			if totalElements != tc.size {
				t.Errorf("bucket中元素总数(%d)与map大小(%d)不一致", totalElements, tc.size)
			}

			t.Logf("测试%s: map大小=%d, bucket数量=%d, 元素总数=%d", tc.name, tc.size, bucketCount, totalElements)
		})
	}
}

// TestMapInspector_RealWorldBehavior 说明Go map在真实环境中的扩容行为
func TestMapInspector_RealWorldBehavior(t *testing.T) {
	t.Log("=== Go map真实扩容行为说明 ===")
	t.Log("在真实的Go运行时中，map的bucket数量会根据元素数量动态变化：")
	t.Log("1. 初始时，map通常有1个bucket（B=0, 2^0=1）")
	t.Log("2. 当元素数量达到负载因子阈值时，会触发扩容")
	t.Log("3. 扩容时bucket数量会翻倍（B增加1，bucket数量变为2^B）")
	t.Log("4. 例如：")
	t.Log("   - B=0时，1个bucket")
	t.Log("   - B=1时，2个buckets")
	t.Log("   - B=2时，4个buckets")
	t.Log("   - B=3时，8个buckets")
	t.Log("   - B=4时，16个buckets")
	t.Log("   ...")
	t.Log("5. 我们的简化实现假设默认有8个buckets，用于教育目的")
	
	// 创建不同大小的map来演示元素分布
	smallMap := make(map[int]int)
	mediumMap := make(map[int]int)
	largeMap := make(map[int]int)
	
	// 填充maps
	for i := 0; i < 10; i++ {
		smallMap[i] = i
	}
	for i := 0; i < 100; i++ {
		mediumMap[i] = i
	}
	for i := 0; i < 1000; i++ {
		largeMap[i] = i
	}
	
	inspector := NewMapInspector()
	
	// 展示元素在buckets中的分布
	t.Log("\n=== 元素在buckets中的分布情况 ===")
	
	showDistribution := func(name string, m map[int]int) {
		buckets, _ := inspector.InspectAllMapBuckets(m)
		t.Logf("%s (大小: %d):", name, len(m))
		for i, bucket := range buckets {
			if len(bucket) > 0 {
				t.Logf("  Bucket %d: %d个元素", i, len(bucket))
			}
		}
	}
	
	showDistribution("小map", smallMap)
	showDistribution("中等map", mediumMap)
	showDistribution("大map", largeMap)
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