// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"
	"errors"

	goredis "github.com/redis/go-redis/v9"
	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/pay_gate/internal/config"
	"github.com/starslipay/pay_gate/internal/middleware"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type ServiceContext struct {
	Config          config.Config
	UserMgr         user_mgr_pb.UserMgrClient
	AccountMgr      account_mgr_pb.AccountMgrClient
	TradeItg        trade_itg_pb.TradeItgClient
	OrderMgr        order_mgr_pb.OrderMgrClient
	Redis           *goredis.Client
	AuthInterceptor rest.Middleware
	RateLimiter     rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	userMgrClient := user_mgr_pb.NewUserMgrClient(zrpc.MustNewClient(c.UserMgrRpcConfig).Conn())
	accountMgrClient := account_mgr_pb.NewAccountMgrClient(zrpc.MustNewClient(c.AccountMgrRpcConfig).Conn())
	tradeItgClient := trade_itg_pb.NewTradeItgClient(zrpc.MustNewClient(c.TradeItgRpcConfig).Conn())
	orderMgrClient := order_mgr_pb.NewOrderMgrClient(zrpc.MustNewClient(c.OrderMgrRpcConfig).Conn())

	authInterceptor := middleware.NewAuthInterceptorMiddleware(&c)

	// 令牌桶限流: 通过哨兵发现并连接当前 Redis 主节点, 主节点故障由哨兵自动切换
	redisClient := goredis.NewFailoverClient(&goredis.FailoverOptions{
		MasterName:    c.Redis.Master,
		SentinelAddrs: c.Redis.Addrs,
		Password:      c.Redis.Password,
		DB:            c.Redis.DB,
	})
	// 裸 go-redis 客户端不会自动产生 trace span(go-zero 只在自身 core/stores/redis 包内挂埋点 hook),
	// 挂上自定义 hook 后, 每次 Redis 命令都会以子 span 形式出现在请求链路中
	redisClient.AddHook(redisTraceHook{})
	rateLimiter := middleware.NewRateLimitMiddleware(redisClient, c.RateLimit)

	return &ServiceContext{
		Config:          c,
		UserMgr:         userMgrClient,
		AccountMgr:      accountMgrClient,
		TradeItg:        tradeItgClient,
		OrderMgr:        orderMgrClient,
		Redis:           redisClient,
		AuthInterceptor: authInterceptor.Handle,
		RateLimiter:     rateLimiter.Handle,
	}
}

// redisTraceHook 为裸 go-redis 客户端补齐 OpenTelemetry 埋点。
// go-zero 仅在自身 core/stores/redis 包内挂载埋点 hook(生成名为 "redis" 的 client span),
// 而本项目因需要哨兵支持直接使用 go-redis, 故在此复刻同样的行为。
// span 通过 trace.TracerFromContext(ctx) 创建, 会自动挂到当前请求的 span 之下。
type redisTraceHook struct{}

// DialHook 不需要埋点, 直接透传
func (redisTraceHook) DialHook(next goredis.DialHook) goredis.DialHook {
	return next
}

// ProcessHook 单条命令: 创建 span, 记录命令名(如 GET/SET), 结束时回填错误与状态
func (redisTraceHook) ProcessHook(next goredis.ProcessHook) goredis.ProcessHook {
	return func(ctx context.Context, cmd goredis.Cmder) error {
		ctx, span := trace.TracerFromContext(ctx).Start(ctx, "redis."+cmd.Name(),
			oteltrace.WithSpanKind(oteltrace.SpanKindClient),
			oteltrace.WithAttributes(attribute.StringSlice("redis.cmds", []string{cmd.Name()})))
		err := next(ctx, cmd)
		endRedisSpan(span, err)
		return err
	}
}

// ProcessPipelineHook 批量命令(pipeline): 与单条命令同样埋点, 汇总所有命令名
func (redisTraceHook) ProcessPipelineHook(next goredis.ProcessPipelineHook) goredis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []goredis.Cmder) error {
		names := make([]string, 0, len(cmds))
		for _, cmd := range cmds {
			names = append(names, cmd.Name())
		}
		ctx, span := trace.TracerFromContext(ctx).Start(ctx, "redis.pipeline",
			oteltrace.WithSpanKind(oteltrace.SpanKindClient),
			oteltrace.WithAttributes(attribute.StringSlice("redis.cmds", names)))
		err := next(ctx, cmds)
		endRedisSpan(span, err)
		return err
	}
}

// endRedisSpan 结束 span: key 不存在(goredis.Nil)属正常未命中, 记为成功; 其余错误记为失败
func endRedisSpan(span oteltrace.Span, err error) {
	defer span.End()
	if err == nil || errors.Is(err, goredis.Nil) {
		span.SetStatus(codes.Ok, "")
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
