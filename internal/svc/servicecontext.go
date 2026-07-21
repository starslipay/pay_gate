// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"net/http"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/pay_gate/internal/config"
	"github.com/starslipay/pay_gate/internal/middleware"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config              config.Config
	UserMgrRpcClient    user_mgr_pb.UserMgrClient
	AccountMgrRpcClient account_mgr_pb.AccountMgrClient
	TradeItgRpcClient   trade_itg_pb.TradeItgClient
	AuthInterceptor     rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	userMgrClient := user_mgr_pb.NewUserMgrClient(zrpc.MustNewClient(c.UserMgrRpcConfig).Conn())
	accountMgrClient := account_mgr_pb.NewAccountMgrClient(zrpc.MustNewClient(c.AccountMgrRpcConfig).Conn())
	tradeItgClient := trade_itg_pb.NewTradeItgClient(zrpc.MustNewClient(c.TradeItgRpcConfig).Conn())

	authInterceptor := middleware.NewAuthInterceptorMiddleware(userMgrClient)

	return &ServiceContext{
		Config:              c,
		UserMgrRpcClient:    userMgrClient,
		AccountMgrRpcClient: accountMgrClient,
		TradeItgRpcClient:   tradeItgClient,
		AuthInterceptor: func(next http.HandlerFunc) http.HandlerFunc {
			return authInterceptor.Handle(next)
		},
	}
}
