package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/starslipay/pay_gate/internal/config"
	"github.com/starslipay/pay_gate/internal/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

const (
	// rateLimitKeyPrefix 限流器在 Redis 中的 key 前缀
	rateLimitKeyPrefix = "ratelimit:"
	// redisPingInterval Redis 健康探测间隔
	redisPingInterval = 3 * time.Second
	// redisPingTimeout 单次探测超时
	redisPingTimeout = time.Second
)

// RateLimitMiddleware 接口维度的令牌桶限流中间件。
// 令牌状态存储在 Redis(哨兵高可用), 因此多个网关实例共享同一份限流配额(全局限流)。
//   - 已在 config.RateLimit.Rules 中配置的接口: 使用其各自的 Rate/Burst;
//   - 未配置的接口: 每个接口各自独立享有 config.RateLimit.Default 这份较大配额;
//   - Redis 不可用时: 直接放行(跳过限流), 保证网关可用性。
type RateLimitMiddleware struct {
	client  *goredis.Client
	conf    config.RateLimitConf
	pathCfg map[string]config.RateRule // path -> 规则(含默认), 已配置的接口预先放入

	// 未配置接口按需创建的限流器缓存, key 为接口路径
	limiters sync.Map // map[string]*redisTokenLimiter

	// redisHealthy 缓存 Redis 健康状态(1=健康), 由后台 goroutine 定期刷新,
	// 避免每个请求都 Ping 一次 Redis。
	redisHealthy int32
}

// NewRateLimitMiddleware 构造限流中间件, 预建已配置接口的限流器,
// 并启动后台 goroutine 定期探测 Redis 健康状态。
// client 由 go-redis FailoverClient 创建, 主节点由哨兵自动发现和切换。
func NewRateLimitMiddleware(client *goredis.Client, conf config.RateLimitConf) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		client:  client,
		conf:    conf,
		pathCfg: make(map[string]config.RateRule, len(conf.Rules)),
	}

	// 预建已配置接口的限流器
	for _, r := range conf.Rules {
		m.pathCfg[r.Path] = config.RateRule{Rate: r.Rate, Burst: r.Burst}
		m.limiters.Store(r.Path, m.newLimiter(r.Path, r.Rate, r.Burst))
	}

	// 初始探测一次, 并启动周期探测
	m.refreshRedisHealth()
	go m.monitorRedis()

	return m
}

// Handle 返回 go-zero 中间件函数
func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 未开启限流, 或 Redis 不可用 -> 直接放行
		if !m.conf.Enable || atomic.LoadInt32(&m.redisHealthy) == 0 {
			next(w, r)
			return
		}

		path := r.URL.Path
		limiter := m.limiterFor(path)
		if !limiter.AllowCtx(r.Context()) {
			logx.Errorf("rate limit triggered, path=%s", path)
			w.WriteHeader(http.StatusOK) // 200
			httpx.ErrorCtx(r.Context(), w, xerr.ErrTooManyRequests)
			return
		}

		next(w, r)
	}
}

// limiterFor 取得某接口的限流器: 已配置接口直接命中缓存;
// 未配置接口按默认配额惰性创建(每个接口各自独立一份)。
func (m *RateLimitMiddleware) limiterFor(path string) *redisTokenLimiter {
	if v, ok := m.limiters.Load(path); ok {
		return v.(*redisTokenLimiter)
	}
	// 未配置接口, 用默认配额创建并缓存(LoadOrStore 防并发重复创建)
	limiter := m.newLimiter(path, m.conf.Default.Rate, m.conf.Default.Burst)
	actual, _ := m.limiters.LoadOrStore(path, limiter)
	return actual.(*redisTokenLimiter)
}

// newLimiter 基于 Redis 令牌桶创建限流器, key 按接口路径隔离。
func (m *RateLimitMiddleware) newLimiter(path string, rate, burst int) *redisTokenLimiter {
	key := fmt.Sprintf("%s%s", rateLimitKeyPrefix, path)
	return newRedisTokenLimiter(rate, burst, m.client, key)
}

// monitorRedis 后台周期探测 Redis 健康状态
func (m *RateLimitMiddleware) monitorRedis() {
	ticker := time.NewTicker(redisPingInterval)
	defer ticker.Stop()
	for range ticker.C {
		m.refreshRedisHealth()
	}
}

// refreshRedisHealth 探测一次 Redis 并更新缓存的健康标志
func (m *RateLimitMiddleware) refreshRedisHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()
	if err := m.client.Ping(ctx).Err(); err == nil {
		atomic.StoreInt32(&m.redisHealthy, 1)
	} else {
		atomic.StoreInt32(&m.redisHealthy, 0)
		logx.Error("redis unavailable, rate limit will be skipped")
	}
}
