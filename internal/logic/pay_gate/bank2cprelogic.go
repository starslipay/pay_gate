// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/pay_gate/internal/xerr"
	"github.com/starslipay/trade_itg/trade_itg_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type Bank2c_preLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBank2c_preLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Bank2c_preLogic {
	return &Bank2c_preLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Bank2c_preLogic) Bank2c_pre(req *types.Bank2CPreReq) (resp *types.Bank2CPreRsp, err error) {
	itgRsp, err := l.svcCtx.TradeItgRpcClient.C2CTransferPre(l.ctx, &trade_itg_pb.C2CTransferPreReq{
		BuyerUserId: req.UserId,
	})
	if err != nil {
		return nil, xerr.ParseRPCError(err)
	}
	resp = &types.Bank2CPreRsp{
		UserId:        req.UserId,
		TransactionId: itgRsp.TransactionId,
	}
	return
}
