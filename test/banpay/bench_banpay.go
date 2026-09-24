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

// Concurrency 并发支付协程数（从用户池中随机选用户发起支付）
const Concurrency = 20

// Duration 压测持续时长，时间内一直发起并发支付请求
const Duration = 60 * time.Second

// Password 注册用户使用的支付密码
const Password = "123456"

// RechargeAmount 每个用户充值金额（单位:分，1000000分=1万元）
const RechargeAmount = 1000000

// PayAmount 每次支付金额（单位:分，1分，确保余额可支撑大量支付请求）
const PayAmount = 1

// Merchants 预置商户ID列表
var Merchants = []string{"2000000000", "3000000000", "4000000000", "5000000000", "6000000000"}

// HTTPTimeout HTTP 客户端超时
const HTTPTimeout = 10 * time.Second

// ===================== 数据结构 =====================

type regUserReq struct {
	UserId   string `json:"user_id,omitempty"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
	Gender   int32  `json:"gender,omitempty"`
	Age      int32  `json:"age,omitempty"`
	Address  string `json:"address,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	IdType   int32  `json:"id_type,omitempty"`
	IdCard   string `json:"id_card,omitempty"`
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

type bank2cPreReq struct {
	UserId string `json:"user_id,omitempty"`
}

type bank2cDoReq struct {
	TransactionId string `json:"transaction_id,omitempty"`
	UserId        string `json:"user_id,omitempty"`
	BankType      int32  `json:"bank_type,omitempty"`
	Amount        int64  `json:"amount,omitempty"`
	VerifyType    int32  `json:"verify_type,omitempty"`
	Password      string `json:"password,omitempty"`
	Memo          string `json:"memo,omitempty"`
}

type payPreReq struct {
	UserId     string `json:"user_id,omitempty"`
	MerchantId string `json:"merchant_id,omitempty"`
}

type banPayReq struct {
	TransactionId string `json:"transaction_id,omitempty"`
	OutOrderNo    string `json:"out_order_no,omitempty"`
	MerchantId    string `json:"merchant_id,omitempty"`
	UserId        string `json:"user_id,omitempty"`
	Amount        int64  `json:"amount,omitempty"`
	VerifyType    int32  `json:"verify_type,omitempty"`
	Password      string `json:"password,omitempty"`
	Memo          string `json:"memo,omitempty"`
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
		var rsp apiRsp
		if err := json.Unmarshal(data, &rsp); err != nil {
			fmt.Printf("  注册用户 %d 解析失败: %v, body=%s\n", i, err, string(data))
			continue
		}
		if rsp.Code != 0 {
			fmt.Printf("  注册用户 %d 业务失败: code=%d, msg=%s\n", i, rsp.Code, rsp.Msg)
			continue
		}
		users = append(users, userInfo{userId: userId})
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

	// ---------- 第三阶段：充值（确保余额充足） ----------
	fmt.Printf("=== 第三阶段: 为 %d 个用户充值 %d 分 ===\n", len(users), RechargeAmount)
	rechargeOk := 0
	for i := range users {
		// bank2c_pre 获取 transaction_id
		preBody := bank2cPreReq{UserId: users[i].userId}
		ok, data := postAndCheck(client, BaseURL+"/api/pay_gate/bank2c_pre", preBody, nil)
		if !ok {
			fmt.Printf("  用户 %d bank2c_pre 失败: %s\n", i, string(data))
			continue
		}
		var preRsp struct {
			UserId        string `json:"user_id"`
			TransactionId string `json:"transaction_id"`
		}
		json.Unmarshal(data, &preRsp)

		// bank2c_do 充值
		doBody := bank2cDoReq{
			TransactionId: preRsp.TransactionId,
			UserId:        users[i].userId,
			BankType:      1, // 默认银行类型
			Amount:        RechargeAmount,
			VerifyType:    1, // 密码验证
			Password:      Password,
			Memo:          "压测充值",
		}
		ok, data = postAndCheck(client, BaseURL+"/api/pay_gate/bank2c_do", doBody, nil)
		if !ok {
			fmt.Printf("  用户 %d bank2c_do 失败: %s\n", i, string(data))
			continue
		}
		rechargeOk++
	}
	fmt.Printf("  充值成功: %d/%d\n\n", rechargeOk, len(users))

	// ---------- 第四阶段：并发支付（基于时间窗口） ----------
	fmt.Printf("=== 第四阶段: 并发压测支付 (用户池=%d, 并发=%d, 持续=%s) ===\n",
		len(users), Concurrency, Duration)

	var (
		successCnt int64 // ban_pay 成功
		failCnt    int64 // ban_pay 失败
		preFailCnt int64 // pay_pre 失败
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
			seq := 0
			for time.Now().Before(deadline) {
				// 随机选用户和商户
				u := users[rand.Intn(len(users))]
				merchantId := Merchants[rand.Intn(len(Merchants))]
				headers := map[string]string{}
				if u.token != "" {
					headers["UserToken"] = u.token
				}

				// 第一步: pay_re 获取 transaction_id
				preBody := payPreReq{
					UserId:     u.userId,
					MerchantId: merchantId,
				}
				ok, data := postAndCheck(client, BaseURL+"/api/pay_gate/pay_re", preBody, headers)
				if !ok {
					atomic.AddInt64(&preFailCnt, 1)
					continue
				}
				var preRsp struct {
					UserId        string `json:"user_id"`
					TransactionId string `json:"transaction_id"`
				}
				json.Unmarshal(data, &preRsp)

				// 生成唯一商户订单号
				seq++
				outOrderNo := fmt.Sprintf("out_%d_%d_%d", uidPrefix, workerId, seq)

				// 第二步: ban_pay 确认支付
				payBody := banPayReq{
					TransactionId: preRsp.TransactionId,
					OutOrderNo:    outOrderNo,
					MerchantId:    merchantId,
					UserId:        u.userId,
					Amount:        PayAmount,
					VerifyType:    1, // 密码验证
					Password:      Password,
					Memo:          "压测支付",
				}

				reqStart := time.Now()
				ok, _ = postAndCheck(client, BaseURL+"/api/pay_gate/ban_pay", payBody, headers)
				latency := time.Since(reqStart)

				if ok {
					atomic.AddInt64(&successCnt, 1)
				} else {
					atomic.AddInt64(&failCnt, 1)
				}

				local = append(local, latency.Nanoseconds())
			}
			localLatencies[workerId] = local
		}(w)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// 合并各 worker 的本地延迟数据（只统计 ban_pay 延迟）
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
	fmt.Printf("  ban_pay 成功:     %d\n", atomic.LoadInt64(&successCnt))
	fmt.Printf("  ban_pay 失败:     %d\n", atomic.LoadInt64(&failCnt))
	fmt.Printf("  pay_pre 失败:     %d\n", atomic.LoadInt64(&preFailCnt))
	fmt.Printf("  ban_pay 总请求数: %d\n", totalReq)
	fmt.Printf("  总耗时:           %.2f s\n", totalTime.Seconds())
	fmt.Printf("  QPS:              %.2f\n", float64(totalReq)/totalTime.Seconds())

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
