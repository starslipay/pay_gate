// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"net/http"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/pay_gate/internal/config"
	"github.com/starslipay/pay_gate/internal/middleware"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
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
}

func NewServiceContext(c config.Config) *ServiceContext {
	userMgrClient := user_mgr_pb.NewUserMgrClient(zrpc.MustNewClient(c.UserMgrRpcConfig).Conn())
	accountMgrClient := account_mgr_pb.NewAccountMgrClient(zrpc.MustNewClient(c.AccountMgrRpcConfig).Conn())
	tradeItgClient := trade_itg_pb.NewTradeItgClient(zrpc.MustNewClient(c.TradeItgRpcConfig).Conn())
	orderMgrClient := order_mgr_pb.NewOrderMgrClient(zrpc.MustNewClient(c.OrderMgrRpcConfig).Conn())

	authInterceptor := middleware.NewAuthInterceptorMiddleware(&c)

	return &ServiceContext{
		Config:     c,
		UserMgr:    userMgrClient,
		AccountMgr: accountMgrClient,
		TradeItg:   tradeItgClient,
		OrderMgr:   orderMgrClient,
		AuthInterceptor: func(next http.HandlerFunc) http.HandlerFunc {
			return authInterceptor.Handle(next)
		},
	}
}
