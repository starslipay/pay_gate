// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	UserMgrRpcConfig    zrpc.RpcClientConf
	AccountMgrRpcConfig zrpc.RpcClientConf
	TradeItgRpcConfig   zrpc.RpcClientConf
	OrderMgrRpcConfig   zrpc.RpcClientConf
	TokenExpireTime     int64

	// Redis 令牌桶限流依赖的 Redis, 令牌状态存 Redis 实现多网关实例全局共享
	Redis redis.RedisConf
	// RateLimit 接口维度限流配置
	RateLimit RateLimitConf
}

// RateLimitConf 网关限流配置
type RateLimitConf struct {
	// Enable 是否开启限流, 默认开启
	Enable bool `json:",default=true"`
	// Default 未单独配置的接口使用的默认配额(每个接口各自独立享有该配额)
	Default RateRule
	// Rules 按接口路径单独配置的限流规则
	Rules []RatePathRule `json:",optional"`
}

// RateRule 令牌桶参数
type RateRule struct {
	// Rate 每秒补充的令牌数(稳态 QPS)
	Rate int
	// Burst 桶容量(允许的瞬时突发量)
	Burst int
}

// RatePathRule 某个接口路径的限流规则
type RatePathRule struct {
	Path  string
	Rate  int
	Burst int
}
