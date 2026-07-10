// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/pay_gate/internal/config"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config              config.Config
	UserMgrRpcClient    user_mgr_pb.UserMgrClient
	AccountMgrRpcClient account_mgr_pb.AccountMgrClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:              c,
		UserMgrRpcClient:    user_mgr_pb.NewUserMgrClient(zrpc.MustNewClient(c.UserMgrRpcConfig).Conn()),
		AccountMgrRpcClient: account_mgr_pb.NewAccountMgrClient(zrpc.MustNewClient(c.AccountMgrRpcConfig).Conn()),
	}
}
