package pay_gate

import (
	"context"

	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/paycomm/xerror"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryRechargeWithdrawOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询充值提现单
func NewQueryRechargeWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRechargeWithdrawOrderLogic {
	return &QueryRechargeWithdrawOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryRechargeWithdrawOrderLogic) QueryRechargeWithdrawOrder(req *types.QueryRechargeWithdrawOrderReq) (resp *types.QueryRechargeWithdrawOrderRsp, err error) {
	rpcResp, err := l.svcCtx.OrderMgr.QueryRechargeWithdrawOrder(l.ctx, &order_mgr_pb.QueryRechargeWithdrawOrderReq{
		TransactionId: req.TransactionId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "OrderMgr.QueryRechargeWithdrawOrder")
	}

	orderInfo := rpcResp.OrderInfo
	return &types.QueryRechargeWithdrawOrderRsp{
		TransactionId:        orderInfo.TransactionId,
		ChannelTransactionId: orderInfo.ChannelTransactionId,
		UserId:               orderInfo.UserId,
		Uid:                  orderInfo.Uid,
		BizType:              orderInfo.BizType,
		BankType:             orderInfo.BankType,
		Amount:               orderInfo.Amount,
		CurType:              orderInfo.CurType,
		TradeState:           orderInfo.TradeState,
		Memo:                 orderInfo.Memo,
		CreateTime:           orderInfo.CreateTime,
		UpdateTime:           orderInfo.UpdateTime,
	}, nil
}
