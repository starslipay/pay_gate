package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type C2CTransferPreReq struct {
	BuyerUserId string `json:"buyer_user_id"`
}

// 错误分类常量
const (
	ErrMarshal      = "JSON序列化失败"
	ErrRequest      = "HTTP请求失败"
	ErrStatusNot200 = "HTTP状态码非200"
	ErrReadBody     = "读取响应体失败"
)

// 每类错误最多保存的样本数量
const maxSamplesPerCategory = 5

type Stats struct {
	successCount int64
	failureCount int64
	totalLatency time.Duration
	minLatency   time.Duration
	maxLatency   time.Duration
	// 错误分类统计
	errorStats map[string]int64
	// 错误样本（每类保留前N条）
	errorSamples map[string][]string
	mu           sync.Mutex
}

func newStats() *Stats {
	return &Stats{
		errorStats:   make(map[string]int64),
		errorSamples: make(map[string][]string),
	}
}

func (s *Stats) addSuccess(latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.successCount++
	s.totalLatency += latency
	if s.minLatency == 0 || latency < s.minLatency {
		s.minLatency = latency
	}
	if latency > s.maxLatency {
		s.maxLatency = latency
	}
}

// addFailure 记录失败请求，category 为错误分类，sample 为错误样本信息
func (s *Stats) addFailure(category, sample string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failureCount++
	s.errorStats[category]++

	// 限制每类错误的样本数量
	if len(s.errorSamples[category]) < maxSamplesPerCategory {
		s.errorSamples[category] = append(s.errorSamples[category], sample)
	}
}

func main() {
	var (
		targetURL    string
		concurrency  int
		qps          int
		totalRequest int
	)

	flag.StringVar(&targetURL, "url", "http://127.0.0.1:30888", "目标服务地址")
	flag.IntVar(&concurrency, "c", 100, "并发数")
	flag.IntVar(&qps, "qps", 10000, "每秒请求数")
	flag.IntVar(&totalRequest, "n", 20000, "总请求数")
	flag.Parse()

	apiURL := fmt.Sprintf("%s/api/pay_gate/c2c_transfer_pre", targetURL)

	fmt.Printf("=== 压测配置 ===\n")
	fmt.Printf("目标地址: %s\n", apiURL)
	fmt.Printf("并发数: %d\n", concurrency)
	fmt.Printf("每秒请求数: %d\n", qps)
	fmt.Printf("总请求数: %d\n", totalRequest)
	fmt.Println("===============")

	stats := newStats()
	sem := make(chan struct{}, concurrency)
	wg := &sync.WaitGroup{}

	ticker := time.NewTicker(time.Second / time.Duration(qps))
	defer ticker.Stop()

	startTime := time.Now()
	requestID := 0

	for requestID < totalRequest {
		select {
		case <-ticker.C:
			sem <- struct{}{}
			wg.Add(1)
			go func(id int) {
				defer func() {
					<-sem
					wg.Done()
				}()

				client := &http.Client{
					Timeout: 30 * time.Second,
					Transport: &http.Transport{
						MaxIdleConns:        100,
						MaxIdleConnsPerHost: 100,
						IdleConnTimeout:     90 * time.Second,
					},
				}

				reqBody := C2CTransferPreReq{BuyerUserId: "51"}
				jsonData, err := json.Marshal(reqBody)
				if err != nil {
					stats.addFailure(ErrMarshal, fmt.Sprintf("requestID=%d err=%v", id, err))
					return
				}

				reqStart := time.Now()
				resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
				latency := time.Since(reqStart)

				if err != nil {
					stats.addFailure(ErrRequest, fmt.Sprintf("requestID=%d latency=%v err=%v", id, latency, err))
					return
				}

				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					body, _ := ioutil.ReadAll(resp.Body)
					stats.addFailure(ErrStatusNot200, fmt.Sprintf("requestID=%d status=%d body=%s", id, resp.StatusCode, string(body)))
					return
				}

				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					stats.addFailure(ErrReadBody, fmt.Sprintf("requestID=%d err=%v", id, err))
					return
				}

				fmt.Printf("Response Body: %s\n", data)
				stats.addSuccess(latency)
			}(requestID)
			requestID++
		}
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	stats.mu.Lock()
	successCount := stats.successCount
	failureCount := stats.failureCount
	totalLatency := stats.totalLatency
	minLatency := stats.minLatency
	maxLatency := stats.maxLatency
	errorStats := stats.errorStats
	errorSamples := stats.errorSamples
	stats.mu.Unlock()

	totalCount := successCount + failureCount
	successRate := float64(successCount) / float64(totalCount) * 100
	failureRate := float64(failureCount) / float64(totalCount) * 100

	var avgLatency time.Duration
	if successCount > 0 {
		avgLatency = totalLatency / time.Duration(successCount)
	}

	qpsResult := float64(totalCount) / totalTime.Seconds()

	fmt.Println("\n=== 压测结果 ===")
	fmt.Printf("总请求数: %d\n", totalCount)
	fmt.Printf("成功数: %d\n", successCount)
	fmt.Printf("失败数: %d\n", failureCount)
	fmt.Printf("成功率: %.2f%%\n", successRate)
	fmt.Printf("失败率: %.2f%%\n", failureRate)
	fmt.Printf("总耗时: %.2f秒\n", totalTime.Seconds())
	fmt.Printf("QPS: %.2f\n", qpsResult)
	if successCount > 0 {
		fmt.Printf("平均耗时: %.2f毫秒\n", float64(avgLatency)/float64(time.Millisecond))
		fmt.Printf("最小耗时: %.2f毫秒\n", float64(minLatency)/float64(time.Millisecond))
		fmt.Printf("最大耗时: %.2f毫秒\n", float64(maxLatency)/float64(time.Millisecond))
	}
	fmt.Println("===============")

	// 打印错误分类统计
	if failureCount > 0 {
		fmt.Println("\n=== 失败错误分类统计 ===")
		fmt.Printf("%-25s %-10s %-10s\n", "错误类型", "数量", "占比")
		fmt.Println("---------------------------------------------")
		for category, count := range errorStats {
			rate := float64(count) / float64(failureCount) * 100
			fmt.Printf("%-25s %-10d %.2f%%\n", category, count, rate)
		}
		fmt.Println("=============================================")

		// 打印每类错误的样本信息
		fmt.Println("\n=== 失败错误样本（每类最多5条） ===")
		for category, samples := range errorSamples {
			fmt.Printf("\n[%s] 共 %d 条样本:\n", category, len(samples))
			for i, sample := range samples {
				fmt.Printf("  %d. %s\n", i+1, sample)
			}
		}
		fmt.Println("=============================================")
	}
}
