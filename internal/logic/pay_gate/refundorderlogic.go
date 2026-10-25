package pay_gate

import (
	"context"

	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_itg/trade_itg_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefundOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 退款
func NewRefundOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundOrderLogic {
	return &RefundOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefundOrderLogic) RefundOrder(req *types.RefundOrderReq) (resp *types.RefundOrderRsp, err error) {
	rpcResp, err := l.svcCtx.TradeItg.RefundOrder(l.ctx, &trade_itg_pb.RefundOrderReq{
		TransactionId: req.TransactionId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "退款")
	}

	return &types.RefundOrderRsp{
		TransactionId: rpcResp.TransactionId,
	}, nil
}
