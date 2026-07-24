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

type C2bank_preLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewC2bank_preLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2bank_preLogic {
	return &C2bank_preLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *C2bank_preLogic) C2bank_pre(req *types.C2BankPreReq) (resp *types.C2BankPreRsp, err error) {
	itgRsp, err := l.svcCtx.TradeItgRpcClient.C2BankPre(l.ctx, &trade_itg_pb.C2BankPreReq{
		UserId: req.UserId,
	})
	if err != nil {
		return nil, xerr.ParseRPCError(err)
	}
	resp = &types.C2BankPreRsp{
		UserId:        req.UserId,
		TransactionId: itgRsp.TransactionId,
	}
	return
}
