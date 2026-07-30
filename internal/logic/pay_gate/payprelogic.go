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

type Pay_preLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPay_preLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Pay_preLogic {
	return &Pay_preLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Pay_preLogic) Pay_pre(req *types.PayPreReq) (resp *types.PayPreRsp, err error) {
	payPreRsp, err := l.svcCtx.TradeItg.PayPre(l.ctx, &trade_itg_pb.PayPreReq{
		UserId:     req.UserId,
		MerchantId: req.MerchantId,
	})
	if err != nil {
		return nil, xerr.HandleRPCError(err, "TradeItg.PayPre")
	}
	return &types.PayPreRsp{
		UserId:        payPreRsp.UserId,
		TransactionId: payPreRsp.TransactionId,
	}, nil
}
