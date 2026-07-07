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

type GetUserInfoReq struct {
	UserId string `json:"user_id"`
}

type Stats struct {
	successCount int64
	failureCount int64
	totalLatency time.Duration
	minLatency   time.Duration
	maxLatency   time.Duration
	mu           sync.Mutex
}

func (s *Stats) addResult(success bool, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if success {
		s.successCount++
		s.totalLatency += latency
		if s.minLatency == 0 || latency < s.minLatency {
			s.minLatency = latency
		}
		if latency > s.maxLatency {
			s.maxLatency = latency
		}
	} else {
		s.failureCount++
	}
}

func main() {
	var (
		targetURL    string
		concurrency  int
		qps          int
		totalRequest int
	)

	flag.StringVar(&targetURL, "url", "http://172.30.40.184:30888", "目标服务地址")
	flag.IntVar(&concurrency, "c", 10, "并发数")
	flag.IntVar(&qps, "qps", 100, "每秒请求数")
	flag.IntVar(&totalRequest, "n", 2000, "总请求数")
	flag.Parse()

	apiURL := fmt.Sprintf("%s/api/pay_gate/get_user_info", targetURL)

	fmt.Printf("=== 压测配置 ===\n")
	fmt.Printf("目标地址: %s\n", apiURL)
	fmt.Printf("并发数: %d\n", concurrency)
	fmt.Printf("每秒请求数: %d\n", qps)
	fmt.Printf("总请求数: %d\n", totalRequest)
	fmt.Println("===============")

	stats := &Stats{}
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

				reqBody := GetUserInfoReq{UserId: fmt.Sprintf("test_user_%d", id)}
				jsonData, err := json.Marshal(reqBody)
				if err != nil {
					stats.addResult(false, 0)
					return
				}

				reqStart := time.Now()
				resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
				latency := time.Since(reqStart)

				if err != nil {
					stats.addResult(false, latency)
					return
				}

				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					stats.addResult(false, latency)
					return
				}

				data, err := ioutil.ReadAll(resp.Body)
				fmt.Printf("Response Body: %s\n", data)
				if err != nil {
					stats.addResult(false, latency)
					return
				}

				stats.addResult(true, latency)
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
}
