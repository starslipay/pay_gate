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

type Get_order_infoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGet_order_infoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Get_order_infoLogic {
	return &Get_order_infoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Get_order_infoLogic) Get_order_info(req *types.GetOrderInfoReq) (resp *types.GetOrderInfoRsp, err error) {
	queryResp, err := l.svcCtx.OrderMgr.QueryOrder(l.ctx, &order_mgr_pb.QueryOrderReq{
		TransactionId: req.TransactionId,
	})
	if err != nil {
		return nil, xerr.HandleRPCError(err, "OrderMgr.QueryOrder")
	}
	resp = &types.GetOrderInfoRsp{
		TransactionId: queryResp.OrderInfo.TransactionId,
		OutOrderNo:    queryResp.OrderInfo.OutOrderNo,
		MerchantId:    queryResp.OrderInfo.MerchantId,
		MerchantName:  queryResp.OrderInfo.MerchantName,
		UserId:        queryResp.OrderInfo.UserId,
		Amount:        queryResp.OrderInfo.Amount,
		PayTime:       queryResp.OrderInfo.PayTime,
	}
	return
}
