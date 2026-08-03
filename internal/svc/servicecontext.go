// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/pay_gate/internal/config"
	"github.com/starslipay/pay_gate/internal/middleware"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config          config.Config
	UserMgr         user_mgr_pb.UserMgrClient
	AccountMgr      account_mgr_pb.AccountMgrClient
	TradeItg        trade_itg_pb.TradeItgClient
	OrderMgr        order_mgr_pb.OrderMgrClient
	AuthInterceptor rest.Middleware
	RateLimiter     rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	userMgrClient := user_mgr_pb.NewUserMgrClient(zrpc.MustNewClient(c.UserMgrRpcConfig).Conn())
	accountMgrClient := account_mgr_pb.NewAccountMgrClient(zrpc.MustNewClient(c.AccountMgrRpcConfig).Conn())
	tradeItgClient := trade_itg_pb.NewTradeItgClient(zrpc.MustNewClient(c.TradeItgRpcConfig).Conn())
	orderMgrClient := order_mgr_pb.NewOrderMgrClient(zrpc.MustNewClient(c.OrderMgrRpcConfig).Conn())

	authInterceptor := middleware.NewAuthInterceptorMiddleware(&c)

	// 令牌桶限流: 令牌状态存 Redis, 多网关实例全局共享配额
	redisStore := redis.MustNewRedis(c.Redis)
	rateLimiter := middleware.NewRateLimitMiddleware(redisStore, c.RateLimit)

	return &ServiceContext{
		Config:          c,
		UserMgr:         userMgrClient,
		AccountMgr:      accountMgrClient,
		TradeItg:        tradeItgClient,
		OrderMgr:        orderMgrClient,
		AuthInterceptor: authInterceptor.Handle,
		RateLimiter:     rateLimiter.Handle,
	}
}
