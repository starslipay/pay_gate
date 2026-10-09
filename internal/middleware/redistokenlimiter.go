package middleware

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	xrate "golang.org/x/time/rate"
)

// tokenLuaScript 与 go-zero core/limit/tokenscript.lua 完全一致的令牌桶算法,
// 保证 Redis 中的 key 结构和限流语义与之前单机版本相同。
const tokenLuaScript = `
-- KEYS[1] as tokens_key
-- KEYS[2] as timestamp_key
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local fill_time = capacity/rate
local ttl = math.floor(fill_time*2)
local last_tokens = tonumber(redis.call("get", KEYS[1]))
if last_tokens == nil then
    last_tokens = capacity
end

local last_refreshed = tonumber(redis.call("get", KEYS[2]))
if last_refreshed == nil then
    last_refreshed = 0
end

local delta = math.max(0, now-last_refreshed)
local filled_tokens = math.min(capacity, last_tokens+(delta*rate))
local allowed = filled_tokens >= requested
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - requested
end

redis.call("setex", KEYS[1], ttl, new_tokens)
redis.call("setex", KEYS[2], ttl, now)

return allowed
`

var (
	tokenScript     = goredis.NewScript(tokenLuaScript)
	tokenFormat     = "{%s}.tokens"
	timestampFormat = "{%s}.ts"
	luaPingInterval = 100 * time.Millisecond
)

// redisTokenLimiter 基于 go-redis(哨兵 FailoverClient) 的令牌桶限流器。
// 主节点由哨兵自动发现和切换; Redis 不可用时退化为进程内 x/time/rate 兜底。
type redisTokenLimiter struct {
	rate         int
	burst        int
	client       *goredis.Client
	tokenKey     string
	timestampKey string

	rescueLock     sync.Mutex
	redisAlive     uint32
	monitorStarted bool
	rescueLimiter  *xrate.Limiter
}

// newRedisTokenLimiter 创建限流器, key 用于在 Redis 中隔离不同接口。
func newRedisTokenLimiter(rate, burst int, client *goredis.Client, key string) *redisTokenLimiter {
	return &redisTokenLimiter{
		rate:         rate,
		burst:        burst,
		client:       client,
		tokenKey:     fmt.Sprintf(tokenFormat, key),
		timestampKey: fmt.Sprintf(timestampFormat, key),
		redisAlive:   1,
		rescueLimiter: xrate.NewLimiter(
			xrate.Every(time.Second/time.Duration(rate)), burst),
	}
}

// AllowCtx 判断当前是否允许 1 个请求通过。
func (l *redisTokenLimiter) AllowCtx(ctx context.Context) bool {
	return l.reserveN(ctx, time.Now(), 1)
}

func (l *redisTokenLimiter) reserveN(ctx context.Context, now time.Time, n int) bool {
	if atomic.LoadUint32(&l.redisAlive) == 0 {
		return l.rescueLimiter.AllowN(now, n)
	}

	resp, err := tokenScript.Run(ctx, l.client,
		[]string{l.tokenKey, l.timestampKey},
		strconv.Itoa(l.rate),
		strconv.Itoa(l.burst),
		strconv.FormatInt(now.Unix(), 10),
		strconv.Itoa(n),
	).Result()

	// Lua boolean false -> nil bulk reply
	if errors.Is(err, goredis.Nil) {
		return false
	}
	// 请求上下文超时/取消, 直接拒绝
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		logx.Errorf("fail to use rate limiter: %s", err)
		return false
	}
	if err != nil {
		logx.Errorf("fail to use rate limiter: %s, use in-process limiter for rescue", err)
		l.startMonitor()
		return l.rescueLimiter.AllowN(now, n)
	}

	code, ok := resp.(int64)
	if !ok {
		logx.Errorf("fail to eval redis script: %v, use in-process limiter for rescue", resp)
		l.startMonitor()
		return l.rescueLimiter.AllowN(now, n)
	}

	// Lua boolean true -> integer reply with value of 1
	return code == 1
}

// startMonitor Redis 出错时启动后台探测, 恢复后重新启用 Redis 限流。
func (l *redisTokenLimiter) startMonitor() {
	l.rescueLock.Lock()
	defer l.rescueLock.Unlock()

	if l.monitorStarted {
		return
	}

	l.monitorStarted = true
	atomic.StoreUint32(&l.redisAlive, 0)

	go l.waitForRedis()
}

func (l *redisTokenLimiter) waitForRedis() {
	ticker := time.NewTicker(luaPingInterval)
	defer func() {
		ticker.Stop()
		l.rescueLock.Lock()
		l.monitorStarted = false
		l.rescueLock.Unlock()
	}()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := l.client.Ping(ctx).Err()
		cancel()
		if err == nil {
			atomic.StoreUint32(&l.redisAlive, 1)
			return
		}
	}
}
