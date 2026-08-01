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

type Close_or_supply_orderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClose_or_supply_orderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Close_or_supply_orderLogic {
	return &Close_or_supply_orderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Close_or_supply_orderLogic) Close_or_supply_order(req *types.CloseOrSupplyOrderReq) (resp *types.CloseOrSupplyOrderRsp, err error) {
	closeOrSupplyOrderRsp, err := l.svcCtx.TradeItg.CloseOrSupplyOrder(l.ctx, &trade_itg_pb.CloseOrSupplyOrderReq{
		TransactionId: req.TransactionId,
		OutOrderNo:    req.OutOrderNo,
		UserId:        req.UserId,
		MerchantId:    req.MerchantId,
		Amount:        req.Amount,
	})
	if err != nil {
		return nil, xerr.HandleRPCError(err, "TradeItg.CloseOrSupplyOrder")
	}
	resp = &types.CloseOrSupplyOrderRsp{
		ResultCode:        closeOrSupplyOrderRsp.ResultCode,
		TransactionId:     closeOrSupplyOrderRsp.TransactionId,
		OutOrderNo:        closeOrSupplyOrderRsp.OutOrderNo,
		UserId:            closeOrSupplyOrderRsp.UserId,
		MerchantId:        closeOrSupplyOrderRsp.MerchantId,
		Amount:            closeOrSupplyOrderRsp.Amount,
		PayTime:           closeOrSupplyOrderRsp.PayTime,
		OrderSuccessToken: closeOrSupplyOrderRsp.OrderSuccessToken,
	}
	return
}
