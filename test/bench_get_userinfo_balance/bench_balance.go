package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// ===================== 可调参数 =====================

// BaseURL pay_gate 服务地址（Docker 映射端口 30888）
const BaseURL = "http://localhost:30888"

// RegUserCount 注册用户数（用户池大小）
const RegUserCount = 10

// Concurrency 并发查询协程数（从用户池中随机选用户发起查询）
const Concurrency = 1000

// Duration 压测持续时长，时间内一直发起并发查询请求
const Duration = 10 * time.Second

// Password 注册用户使用的支付密码
const Password = "123456"

// HTTPTimeout HTTP 客户端超时
const HTTPTimeout = 10 * time.Second

// ===================== 数据结构 =====================

type regUserReq struct {
	UserId   string `json:"user_id,omitempty"`  // 用户ID（必填, 1~64字符）
	Password string `json:"password,omitempty"` // 支付密码（必填, 1~64字符）
	Name     string `json:"name,omitempty"`     // 用户姓名（必填, 1~64字符）
	Gender   int32  `json:"gender,omitempty"`   // 性别: 1男 2女
	Age      int32  `json:"age,omitempty"`      // 年龄（必填, 1~256）
	Address  string `json:"address,omitempty"`  // 联系地址（必填, 1~64字符）
	Phone    string `json:"phone,omitempty"`    // 手机号（必填, 1~64字符）
	Email    string `json:"email,omitempty"`    // 邮箱（必填, 1~64字符）
	IdType   int32  `json:"id_type,omitempty"`  // 证件类型: 1身份证（必填, 1~64）
	IdCard   string `json:"id_card,omitempty"`  // 证件号码（必填, 1~64字符）
}

type regUserRsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		UserId string `json:"user_id"`
	} `json:"data"`
}

type getUserTokenReq struct {
	UserId   string `json:"user_id,omitempty"`
	Password string `json:"password,omitempty"`
}

type getUserTokenRsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		UserId    string `json:"user_id"`
		UserToken string `json:"user_token"`
	} `json:"data"`
}

type getUserBalanceReq struct {
	UserId string `json:"user_id,omitempty"`
}

type getUserBalanceRsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		UserId  string `json:"user_id"`
		Balance int64  `json:"balance"`
		CurType int32  `json:"cur_type"`
	} `json:"data"`
}

// ===================== 工具函数 =====================

func httpPost(client *http.Client, url string, body interface{}, headers map[string]string) ([]byte, int, error) {
	jsonBytes, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return data, resp.StatusCode, nil
}

// ===================== 压测主流程 =====================

type userInfo struct {
	userId string
	token  string
}

func main() {
	maxConn := Concurrency
	if RegUserCount > Concurrency {
		maxConn = RegUserCount
	}
	client := &http.Client{
		Timeout: HTTPTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        maxConn + 10,
			MaxIdleConnsPerHost: maxConn + 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// ---------- 第一阶段：注册用户，构建用户池 ----------
	fmt.Printf("=== 第一阶段: 注册 %d 个用户 ===\n", RegUserCount)
	users := make([]userInfo, 0, RegUserCount)
	uidPrefix := time.Now().UnixMilli()
	for i := 0; i < RegUserCount; i++ {
		userId := fmt.Sprintf("bench_%d_%d", uidPrefix, i)
		body := regUserReq{
			UserId:   userId,
			Password: Password,
			Name:     "压测用户" + strconv.Itoa(i),
			Gender:   1,
			Age:      20 + int32(i%30),
			Address:  "深圳市南山区科技园",
			Phone:    fmt.Sprintf("138%08d", i),
			Email:    fmt.Sprintf("bench%d@test.com", i),
			IdType:   1,
			IdCard:   fmt.Sprintf("4403%014d", i),
		}
		data, _, err := httpPost(client, BaseURL+"/api/pay_gate/reg_user", body, nil)
		if err != nil {
			fmt.Printf("  注册用户 %d 失败: %v\n", i, err)
			continue
		}
		var rsp regUserRsp
		if err := json.Unmarshal(data, &rsp); err != nil {
			fmt.Printf("  注册用户 %d 解析失败: %v, body=%s\n", i, err, string(data))
			continue
		}
		if rsp.Code != 0 {
			fmt.Printf("  注册用户 %d 业务失败: code=%d, msg=%s\n", i, rsp.Code, rsp.Msg)
			continue
		}
		users = append(users, userInfo{userId: rsp.Data.UserId})
	}
	fmt.Printf("  注册成功: %d/%d\n\n", len(users), RegUserCount)

	if len(users) == 0 {
		fmt.Println("没有有效用户，退出压测")
		return
	}

	// ---------- 第二阶段：获取用户 token ----------
	fmt.Printf("=== 第二阶段: 获取 %d 个用户 token ===\n", len(users))
	for i := range users {
		body := getUserTokenReq{
			UserId:   users[i].userId,
			Password: Password,
		}
		data, _, err := httpPost(client, BaseURL+"/api/pay_gate/get_user_token", body, nil)
		if err != nil {
			fmt.Printf("  获取 token %d 失败: %v\n", i, err)
			continue
		}
		var rsp getUserTokenRsp
		if err := json.Unmarshal(data, &rsp); err != nil {
			fmt.Printf("  获取 token %d 解析失败: %v, body=%s\n", i, err, string(data))
			continue
		}
		if rsp.Code != 0 {
			fmt.Printf("  获取 token %d 业务失败: code=%d, msg=%s\n", i, rsp.Code, rsp.Msg)
			continue
		}
		users[i].token = rsp.Data.UserToken
	}
	tokenCount := 0
	for _, u := range users {
		if u.token != "" {
			tokenCount++
		}
	}
	fmt.Printf("  获取 token 成功: %d/%d\n\n", tokenCount, len(users))

	// ---------- 第三阶段：并发查询余额（基于时间窗口） ----------
	fmt.Printf("=== 第三阶段: 并发压测查询余额 (用户池=%d, 并发=%d, 持续=%s) ===\n",
		len(users), Concurrency, Duration)

	var (
		successCnt int64
		failCnt    int64
	)

	startTime := time.Now()
	deadline := startTime.Add(Duration)
	var wg sync.WaitGroup

	// 每个 worker 用本地 slice 收集延迟，避免全局锁竞争
	localLatencies := make([][]int64, Concurrency)
	for i := range localLatencies {
		localLatencies[i] = make([]int64, 0, 1024)
	}

	for w := 0; w < Concurrency; w++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			local := localLatencies[workerId]
			for time.Now().Before(deadline) {
				// 随机选一个用户查询
				u := users[rand.Intn(len(users))]
				body := getUserBalanceReq{UserId: u.userId}
				headers := map[string]string{}
				if u.token != "" {
					headers["UserToken"] = u.token
				}

				reqStart := time.Now()
				respData, _, err := httpPost(client, BaseURL+"/api/pay_gate/get_user_balance_info", body, headers)
				latency := time.Since(reqStart)

				if err != nil {
					atomic.AddInt64(&failCnt, 1)
				} else {
					var rsp getUserBalanceRsp
					if json.Unmarshal(respData, &rsp) == nil && rsp.Code == 0 {
						atomic.AddInt64(&successCnt, 1)
					} else {
						atomic.AddInt64(&failCnt, 1)
					}
				}

				local = append(local, latency.Nanoseconds())
			}
			localLatencies[workerId] = local
		}(w)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// 合并各 worker 的本地延迟数据
	totalReq := atomic.LoadInt64(&successCnt) + atomic.LoadInt64(&failCnt)
	latencies := make([]int64, 0, totalReq)
	var totalLatency int64
	var maxLatency int64
	for _, local := range localLatencies {
		latencies = append(latencies, local...)
		for _, l := range local {
			totalLatency += l
			if l > maxLatency {
				maxLatency = l
			}
		}
	}

	// ---------- 统计结果 ----------
	fmt.Printf("\n=== 压测结果 ===\n")
	fmt.Printf("  总请求数:   %d\n", totalReq)
	fmt.Printf("  成功:       %d\n", atomic.LoadInt64(&successCnt))
	fmt.Printf("  失败:       %d\n", atomic.LoadInt64(&failCnt))
	fmt.Printf("  总耗时:     %.2f s\n", totalTime.Seconds())
	fmt.Printf("  QPS:        %.2f\n", float64(totalReq)/totalTime.Seconds())

	if len(latencies) > 0 {
		avgLatencyMs := float64(totalLatency) / float64(len(latencies)) / 1e6
		maxLatencyMs := float64(maxLatency) / 1e6

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p50 := float64(latencies[len(latencies)*50/100]) / 1e6
		p90 := float64(latencies[len(latencies)*90/100]) / 1e6
		p99 := float64(latencies[len(latencies)*99/100]) / 1e6

		fmt.Printf("  平均延迟:   %.2f ms\n", avgLatencyMs)
		fmt.Printf("  P50 延迟:   %.2f ms\n", p50)
		fmt.Printf("  P90 延迟:   %.2f ms\n", p90)
		fmt.Printf("  P99 延迟:   %.2f ms\n", p99)
		fmt.Printf("  最大延迟:   %.2f ms\n", maxLatencyMs)
	}
}
