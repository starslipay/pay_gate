// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/pay_gate/internal/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type Query_orderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQuery_orderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Query_orderLogic {
	return &Query_orderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Query_orderLogic) Query_order(req *types.QueryOrderReq) (resp *types.QueryOrderRsp, err error) {
	queryResp, err := l.svcCtx.OrderMgr.QueryOrder(l.ctx, &order_mgr_pb.QueryOrderReq{
		TransactionId: req.TransactionId,
	})
	if err != nil {
		return nil, xerr.HandleRPCError(err, "OrderMgr.QueryOrder")
	}
	// 返回 order_mgr 查询到的全部订单信息
	resp = &types.QueryOrderRsp{
		TransactionId:     queryResp.OrderInfo.TransactionId,
		OutOrderNo:        queryResp.OrderInfo.OutOrderNo,
		MerchantId:        queryResp.OrderInfo.MerchantId,
		MerchantUid:       queryResp.OrderInfo.MerchantUid,
		MerchantName:      queryResp.OrderInfo.MerchantName,
		UserId:            queryResp.OrderInfo.UserId,
		Uid:               queryResp.OrderInfo.Uid,
		Amount:            queryResp.OrderInfo.Amount,
		PayTime:           queryResp.OrderInfo.PayTime,
		TradeState:        queryResp.OrderInfo.TradeState,
		OrderSuccessToken: queryResp.OrderInfo.OrderSuccessToken,
	}
	return
}
