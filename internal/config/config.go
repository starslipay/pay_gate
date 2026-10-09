// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
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

	// Redis 哨兵高可用配置, 令牌桶限流状态存 Redis 实现多网关实例全局共享
	Redis RedisConf
	// RateLimit 接口维度限流配置
	RateLimit RateLimitConf

	// AccessLog 请求/响应日志脱敏配置
	AccessLog AccessLogConf
}

// RedisConf Redis 哨兵(Sentinel)高可用配置。
// go-zero 的 RedisConf 仅支持 node/cluster, 不支持哨兵, 故此处自定义,
// 运行时用 go-redis 的 FailoverClient 连接(自动发现并切换主节点)。
type RedisConf struct {
	// Master 哨兵监控的主节点名称, 需与 sentinel.conf 中 "sentinel monitor" 的名称一致
	Master string
	// Addrs 哨兵节点地址列表, go-redis 会依次询问以发现当前主节点
	Addrs []string
	// Password Redis 访问密码(无密码留空)
	Password string `json:",optional"`
	// DB 使用的逻辑库编号, 默认 0
	DB int `json:",default=0"`
}

// AccessLogConf 访问日志配置
type AccessLogConf struct {
	// Enable 是否开启请求/响应日志, 默认开启
	Enable bool `json:",default=true"`
	// SensitiveFields 需要脱敏的字段名列表(不区分大小写)
	SensitiveFields []string `json:",optional"`
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
