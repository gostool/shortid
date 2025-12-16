package main

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gostool/shortid"
)

// ShortLinkService 短链接服务（内存实现，不考虑存储）
type ShortLinkService struct {
	// 使用多个生成器实例，减少锁竞争
	generators []*shortid.DefaultIDGenerator

	// 内存存储（仅用于演示，实际应使用数据库）
	links sync.Map // map[string]string: code -> originalURL

	// 统计信息
	writeCount int64
	readCount  int64
}

// NewShortLinkService 创建短链接服务
func NewShortLinkService() *ShortLinkService {
	// 创建多个生成器实例，每个使用不同的 MachineID
	numGenerators := 10
	generators := make([]*shortid.DefaultIDGenerator, numGenerators)

	for i := 0; i < numGenerators; i++ {
		generators[i] = shortid.NewIDGenerator(shortid.IDGeneratorConfig{
			BusinessType: shortid.BusinessShare, // 'A' 表示分享链接
			EnableDate:   true,                  // 启用日期编码
			DateBase:     "2024-01-01",
			MachineID:    int64(i), // 使用不同的机器ID
		})
	}

	return &ShortLinkService{
		generators: generators,
	}
}

// CreateShortLink 创建短链接
func (s *ShortLinkService) CreateShortLink(originalURL string) (string, error) {
	// 轮询使用不同的生成器，减少锁竞争
	genIdx := atomic.AddInt64(&s.writeCount, 1) % int64(len(s.generators))
	gen := s.generators[genIdx]

	// 生成短码
	shortCode := gen.Generate()

	// 存储映射关系（内存存储，实际应写入数据库）
	s.links.Store(shortCode, originalURL)

	return shortCode, nil
}

// GetOriginalURL 根据短码获取原始URL
func (s *ShortLinkService) GetOriginalURL(shortCode string) (string, bool) {
	atomic.AddInt64(&s.readCount, 1)

	// 从内存获取（实际应从数据库或缓存查询）
	value, ok := s.links.Load(shortCode)
	if !ok {
		return "", false
	}

	originalURL, ok := value.(string)
	return originalURL, ok
}

// BatchCreateShortLinks 批量创建短链接
func (s *ShortLinkService) BatchCreateShortLinks(originalURLs []string) ([]string, error) {
	shortCodes := make([]string, len(originalURLs))

	// 使用单个生成器批量生成，减少锁竞争
	gen := s.generators[0]

	for i, url := range originalURLs {
		shortCode := gen.Generate()
		s.links.Store(shortCode, url)
		shortCodes[i] = shortCode
		atomic.AddInt64(&s.writeCount, 1)
	}

	return shortCodes, nil
}

// GetStats 获取统计信息
func (s *ShortLinkService) GetStats() map[string]int64 {
	return map[string]int64{
		"write_count": atomic.LoadInt64(&s.writeCount),
		"read_count":  atomic.LoadInt64(&s.readCount),
		"total_links": s.getTotalLinks(),
	}
}

func (s *ShortLinkService) getTotalLinks() int64 {
	count := int64(0)
	s.links.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// PerformanceTest 性能测试
func PerformanceTest() {
	service := NewShortLinkService()

	// 测试参数
	writeConcurrency := 100 // 写入并发数
	readConcurrency := 1000 // 读取并发数（读写比 1:10）
	writesPerGoroutine := 10000
	readsPerGoroutine := 100000

	fmt.Println("=== 短链接服务性能测试 ===")
	fmt.Printf("写入并发数: %d, 每个goroutine写入: %d\n", writeConcurrency, writesPerGoroutine)
	fmt.Printf("读取并发数: %d, 每个goroutine读取: %d\n", readConcurrency, readsPerGoroutine)
	fmt.Println()

	var wg sync.WaitGroup

	// 准备一些短码用于读取测试
	testCodes := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		code, _ := service.CreateShortLink(fmt.Sprintf("https://example.com/page/%d", i))
		testCodes[i] = code
	}

	// 写入测试
	writeStart := time.Now()
	for i := 0; i < writeConcurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				url := fmt.Sprintf("https://example.com/page/%d-%d", id, j)
				service.CreateShortLink(url)
			}
		}(i)
	}
	wg.Wait()
	writeDuration := time.Since(writeStart)

	// 读取测试
	readStart := time.Now()
	for i := 0; i < readConcurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < readsPerGoroutine; j++ {
				codeIdx := (id*readsPerGoroutine + j) % len(testCodes)
				service.GetOriginalURL(testCodes[codeIdx])
			}
		}(i)
	}
	wg.Wait()
	readDuration := time.Since(readStart)

	// 统计结果
	totalWrites := int64(writeConcurrency * writesPerGoroutine)
	totalReads := int64(readConcurrency * readsPerGoroutine)

	writeQPS := float64(totalWrites) / writeDuration.Seconds()
	readQPS := float64(totalReads) / readDuration.Seconds()
	totalQPS := writeQPS + readQPS

	fmt.Println("=== 测试结果 ===")
	fmt.Printf("写入测试:\n")
	fmt.Printf("  总写入数: %d\n", totalWrites)
	fmt.Printf("  耗时: %v\n", writeDuration)
	fmt.Printf("  写入QPS: %.0f\n", writeQPS)
	fmt.Printf("  平均延迟: %.2f μs\n", writeDuration.Seconds()*1000000/float64(totalWrites))

	fmt.Printf("\n读取测试:\n")
	fmt.Printf("  总读取数: %d\n", totalReads)
	fmt.Printf("  耗时: %v\n", readDuration)
	fmt.Printf("  读取QPS: %.0f\n", readQPS)
	fmt.Printf("  平均延迟: %.2f μs\n", readDuration.Seconds()*1000000/float64(totalReads))

	fmt.Printf("\n总体性能:\n")
	fmt.Printf("  总QPS: %.0f\n", totalQPS)
	fmt.Printf("  读写比: 1:%.2f\n", readQPS/writeQPS)

	stats := service.GetStats()
	fmt.Printf("\n统计信息:\n")
	fmt.Printf("  总写入数: %d\n", stats["write_count"])
	fmt.Printf("  总读取数: %d\n", stats["read_count"])
	fmt.Printf("  存储链接数: %d\n", stats["total_links"])

	// 评估每天1亿写入的可行性
	fmt.Println("\n=== 可行性评估 ===")
	dailyWrites := int64(100_000_000) // 1亿
	dailyReads := dailyWrites * 10    // 10亿（读写比 1:10）

	writeQPSNeeded := float64(dailyWrites) / 86400.0
	readQPSNeeded := float64(dailyReads) / 86400.0
	totalQPSNeeded := writeQPSNeeded + readQPSNeeded

	fmt.Printf("需求:\n")
	fmt.Printf("  每天写入: %d\n", dailyWrites)
	fmt.Printf("  每天读取: %d\n", dailyReads)
	fmt.Printf("  需要写入QPS: %.0f\n", writeQPSNeeded)
	fmt.Printf("  需要读取QPS: %.0f\n", readQPSNeeded)
	fmt.Printf("  需要总QPS: %.0f\n", totalQPSNeeded)

	fmt.Printf("\n实际性能:\n")
	fmt.Printf("  实际写入QPS: %.0f\n", writeQPS)
	fmt.Printf("  实际读取QPS: %.0f\n", readQPS)
	fmt.Printf("  实际总QPS: %.0f\n", totalQPS)

	writeMargin := writeQPS / writeQPSNeeded
	readMargin := readQPS / readQPSNeeded
	totalMargin := totalQPS / totalQPSNeeded

	fmt.Printf("\n性能余量:\n")
	fmt.Printf("  写入性能余量: %.1f倍\n", writeMargin)
	fmt.Printf("  读取性能余量: %.1f倍\n", readMargin)
	fmt.Printf("  总体性能余量: %.1f倍\n", totalMargin)

	if writeMargin >= 1.0 && readMargin >= 1.0 {
		fmt.Println("\n✅ 结论: 完全可行！性能完全满足需求。")
	} else {
		fmt.Println("\n❌ 结论: 需要优化或扩容。")
	}
}

func main() {
	service := NewShortLinkService()

	fmt.Println("=== 短链接服务示例 ===")
	fmt.Println()

	// 1. 创建短链接
	fmt.Println("1. 创建短链接:")
	originalURL := "https://www.example.com/very/long/url/path/to/resource"
	shortCode, err := service.CreateShortLink(originalURL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   原始URL: %s\n", originalURL)
	fmt.Printf("   短码: %s\n", shortCode)
	fmt.Printf("   短链接: https://short.ly/%s\n", shortCode)
	fmt.Println()

	// 2. 解析短链接
	fmt.Println("2. 解析短链接:")
	retrievedURL, ok := service.GetOriginalURL(shortCode)
	if ok {
		fmt.Printf("   短码: %s\n", shortCode)
		fmt.Printf("   原始URL: %s\n", retrievedURL)
	} else {
		fmt.Printf("   短码不存在: %s\n", shortCode)
	}
	fmt.Println()

	// 3. 批量创建
	fmt.Println("3. 批量创建短链接:")
	urls := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/page3",
	}
	codes, err := service.BatchCreateShortLinks(urls)
	if err != nil {
		log.Fatal(err)
	}
	for i, code := range codes {
		fmt.Printf("   %s -> %s\n", code, urls[i])
	}
	fmt.Println()

	// 4. 性能测试
	// fmt.Println()
	// PerformanceTest()
}
