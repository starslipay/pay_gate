package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ===================== 可调参数 =====================

// BaseURL pay_gate 服务地址（Docker 映射端口 30888）
const BaseURL = "http://localhost:30888"

// RegUserCount 注册用户数（用户池大小）
const RegUserCount = 200

// Concurrency 并发查询协程数（从用户池中循环选用户发起查询）
const Concurrency = 1000

// Duration 压测持续时长，时间内一直发起并发查询请求
const Duration = 60 * time.Second

// Password 注册用户使用的支付密码
const Password = "123456"

// PrepareConcurrency 准备阶段(注册)的最大并发数, 避免瞬时SYN突发打爆Docker Desktop端口转发层
const PrepareConcurrency = 300

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

type apiRsp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type getUserTokenReq struct {
	UserId   string `json:"user_id,omitempty"`
	Password string `json:"password,omitempty"`
}

type getUserBalanceReq struct {
	UserId string `json:"user_id,omitempty"` // 用户ID
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

// initialConcurrency 返回起步并发数（配置并发数的10%，至少1个）
func initialConcurrency() int {
	v := Concurrency / 10
	if v < 1 {
		v = 1
	}
	return v
}

// stepConcurrency 返回第 step 档（从1开始）的目标并发数
// step=1 → 10%, step=2 → 20%, ... step>=10 → 100%
func stepConcurrency(step int) int {
	p := step * 10
	if p > 100 {
		p = 100
	}
	v := Concurrency * p / 100
	if v < 1 {
		v = 1
	}
	return v
}

// postAndCheck 发送请求并检查响应码是否为0（成功），返回 data 字段的原始 JSON
func postAndCheck(client *http.Client, url string, body interface{}, headers map[string]string) (bool, json.RawMessage) {
	respData, _, err := httpPost(client, url, body, headers)
	if err != nil {
		return false, nil
	}
	var rsp apiRsp
	if err := json.Unmarshal(respData, &rsp); err != nil {
		return false, respData
	}
	return rsp.Code == 0, rsp.Data
}

// ===================== 阶梯窗口抽样 =====================

const (
	maxSampleWindows = 64 // 最大采样窗口数(5秒一个窗口, 60s压测=12个, 留足余量)
	samplesPerWindow = 5  // 每个阶梯窗口抽样的 user_id 个数
)

var (
	curWindowIdx    atomic.Int64 // 当前阶梯窗口编号(0起), 控制器每打印一个窗口后递增
	windowSampleCnt [maxSampleWindows]atomic.Int64
	windowSamples   [maxSampleWindows][]string
	windowSampleMu  sync.Mutex
)

// sampleUserId 在当前阶梯窗口内抽样记录 user_id, 每窗口最多 samplesPerWindow 个,
// 仅在未抽满时加锁, 对压测热路径几乎无开销
func sampleUserId(userId string) {
	w := curWindowIdx.Load()
	if w < 0 || w >= maxSampleWindows {
		return
	}
	if windowSampleCnt[w].Load() >= samplesPerWindow {
		return // 本窗口已抽满
	}
	windowSampleMu.Lock()
	if len(windowSamples[w]) < samplesPerWindow {
		windowSamples[w] = append(windowSamples[w], userId)
		windowSampleCnt[w].Add(1)
	}
	windowSampleMu.Unlock()
}

// printWindowSamples 压测结束后打印各阶梯窗口抽样的 user_id, 用于按 user_id 查询日志/trace
func printWindowSamples() {
	last := int(curWindowIdx.Load())
	if last >= maxSampleWindows {
		last = maxSampleWindows - 1
	}
	fmt.Printf("\n=== 各阶梯抽样 user_id(用于查询日志) ===\n")
	for w := 0; w <= last; w++ {
		if len(windowSamples[w]) == 0 {
			continue
		}
		fmt.Printf("  第%d档: %s\n", w+1, strings.Join(windowSamples[w], ", "))
	}
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

	// ---------- 第一阶段：并发注册用户，构建用户池 ----------
	fmt.Printf("=== 第一阶段: 并发注册 %d 个用户 ===\n", RegUserCount)
	uidPrefix := time.Now().UnixMilli()

	// 预分配槽位，各协程按下标写入，避免 append 竞争
	type regResult struct {
		user userInfo
		ok   bool
	}
	results := make([]regResult, RegUserCount)
	var regWg sync.WaitGroup
	regSem := make(chan struct{}, PrepareConcurrency) // 并发闸门
	for i := 0; i < RegUserCount; i++ {
		regSem <- struct{}{} // 超过300个在途时阻塞等待
		regWg.Add(1)
		go func(i int) {
			defer regWg.Done()
			defer func() { <-regSem }()
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
				return
			}
			var rsp apiRsp
			if err := json.Unmarshal(data, &rsp); err != nil {
				fmt.Printf("  注册用户 %d 解析失败: %v, body=%s\n", i, err, string(data))
				return
			}
			if rsp.Code != 0 {
				fmt.Printf("  注册用户 %d 业务失败: code=%d, msg=%s\n", i, rsp.Code, rsp.Msg)
				return
			}
			results[i] = regResult{user: userInfo{userId: userId}, ok: true}
		}(i)
	}
	regWg.Wait()

	// 按原顺序收集注册成功的用户
	users := make([]userInfo, 0, RegUserCount)
	for _, r := range results {
		if r.ok {
			users = append(users, r.user)
		}
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
		var rsp apiRsp
		if err := json.Unmarshal(data, &rsp); err != nil {
			fmt.Printf("  获取 token %d 解析失败: %v\n", i, err)
			continue
		}
		if rsp.Code != 0 {
			fmt.Printf("  获取 token %d 业务失败: code=%d, msg=%s\n", i, rsp.Code, rsp.Msg)
			continue
		}
		// 从 data 中提取 user_token
		var tokenData struct {
			UserId    string `json:"user_id"`
			UserToken string `json:"user_token"`
		}
		json.Unmarshal(rsp.Data, &tokenData)
		users[i].token = tokenData.UserToken
	}
	tokenCount := 0
	for _, u := range users {
		if u.token != "" {
			tokenCount++
		}
	}
	fmt.Printf("  获取 token 成功: %d/%d\n\n", tokenCount, len(users))

	// ---------- 第三阶段：阶梯式并发查询余额（10% → 20% → ... → 100%，每档间隔5秒，实时打印QPS） ----------
	fmt.Printf("=== 第三阶段: 阶梯压测查询余额 (用户池=%d, 目标并发=%d, 持续=%s, 起步并发=%d) ===\n",
		len(users), Concurrency, Duration, initialConcurrency())

	var (
		successCnt int64 // 查询成功
		failCnt    int64 // 查询失败
		latSumNs   int64 // 查询延迟总和(ns), 用于阶梯窗口平均耗时
	)

	startTime := time.Now()
	deadline := startTime.Add(Duration)
	var wg sync.WaitGroup

	// 并发闸门：worker 每轮迭代必须先取到令牌才能发请求；
	// 初始令牌数为并发数的10%，之后每5秒补发下一档，直到100%
	tokens := make(chan struct{}, Concurrency)
	for i := 0; i < initialConcurrency(); i++ {
		tokens <- struct{}{}
	}

	// 阶梯控制器：每5秒提升一档并发，并打印当前窗口QPS/平均耗时/成功失败数
	wg.Add(1)
	go func() {
		defer wg.Done()
		active := initialConcurrency()
		lastDone, lastSucc, lastFail, lastLat := int64(0), int64(0), int64(0), int64(0)
		lastTime := time.Now()
		// printWindow 打印近5秒窗口的QPS、平均耗时、本步与累计成功/失败数
		printWindow := func() {
			now := time.Now()
			curWindowIdx.Add(1) // 当前采样窗口结束, 后续请求落入下一窗口
			succ := atomic.LoadInt64(&successCnt)
			fail := atomic.LoadInt64(&failCnt)
			lat := atomic.LoadInt64(&latSumNs)
			done := succ + fail
			dReq := done - lastDone
			windowQps := float64(dReq) / now.Sub(lastTime).Seconds()
			avgMs := 0.0
			if dReq > 0 {
				avgMs = float64(lat-lastLat) / float64(dReq) / 1e6
			}
			fmt.Printf("  [阶梯] 活跃并发=%d/%d, 近5s QPS=%.2f, 平均耗时=%.1fms, 本步成功=%d 失败=%d, 累计: 成功=%d 失败=%d\n",
				active, Concurrency, windowQps, avgMs, succ-lastSucc, fail-lastFail, succ, fail)
			lastDone, lastSucc, lastFail, lastLat, lastTime = done, succ, fail, lat, now
		}
		for step := 2; ; step++ {
			select {
			case <-time.After(5 * time.Second):
			}
			if !time.Now().Before(deadline) {
				return
			}
			printWindow()
			// 计算下一档并补发令牌
			target := stepConcurrency(step)
			for active < target {
				tokens <- struct{}{}
				active++
			}
			if active >= Concurrency {
				// 已到100%，继续每5秒打印直到压测结束
				for time.Now().Before(deadline) {
					<-time.After(5 * time.Second)
					printWindow()
				}
				return
			}
		}
	}()

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
			userIdx := workerId % len(users) // 当前协程的用户游标, 按 workerId 加初始偏移, 避免所有协程从同一用户齐步走
			for time.Now().Before(deadline) {
				// 取不到令牌时阻塞，等阶梯控制器补发 → 未激活的 worker 不产生请求
				<-tokens
				// 循环从用户数组第一个到最后一个依次取用户，取完从头再来
				u := users[userIdx]
				userIdx = (userIdx + 1) % len(users)
				sampleUserId(u.userId) // 按当前阶梯窗口抽样 user_id, 供压测结束后查日志
				body := getUserBalanceReq{UserId: u.userId}
				headers := map[string]string{}
				if u.token != "" {
					headers["UserToken"] = u.token
				}

				reqStart := time.Now()
				ok, _ := postAndCheck(client, BaseURL+"/api/pay_gate/get_user_balance_info", body, headers)
				latency := time.Since(reqStart)

				if ok {
					atomic.AddInt64(&successCnt, 1)
				} else {
					atomic.AddInt64(&failCnt, 1)
				}

				local = append(local, latency.Nanoseconds())
				atomic.AddInt64(&latSumNs, latency.Nanoseconds())
				tokens <- struct{}{} // 归还令牌，进入下一轮
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
	fmt.Printf("  查询成功:   %d\n", atomic.LoadInt64(&successCnt))
	fmt.Printf("  查询失败:   %d\n", atomic.LoadInt64(&failCnt))
	fmt.Printf("  总请求数:   %d\n", totalReq)
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

	// 打印各阶梯窗口抽样的 user_id, 用于查询日志/trace
	printWindowSamples()
}
