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

type Ban_payLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBan_payLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Ban_payLogic {
	return &Ban_payLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Ban_payLogic) Ban_pay(req *types.BanPayReq) (resp *types.BanPayRsp, err error) {
	banPayRsp, err := l.svcCtx.TradeItg.BanPay(l.ctx, &trade_itg_pb.BanPayReq{
		TransactionId: req.TransactionId,
		OutOrderNo:    req.OutOrderNo,
		UserId:        req.UserId,
		MerchantId:    req.MerchantId,
		Amount:        req.Amount,
		VerifyType:    req.VerifyType,
		Password:      req.Password,
	})
	if err != nil {
		return nil, xerr.HandleRPCError(err, "TradeItg.BanPay")
	}
	return &types.BanPayRsp{
		UserId:            banPayRsp.UserId,
		TransactionId:     banPayRsp.TransactionId,
		OrderSuccessToken: banPayRsp.OrderSuccessToken,
	}, nil
}
