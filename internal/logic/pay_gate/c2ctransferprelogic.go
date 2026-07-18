// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type C2c_transfer_preLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewC2c_transfer_preLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2c_transfer_preLogic {
	return &C2c_transfer_preLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *C2c_transfer_preLogic) C2c_transfer_pre(req *types.C2CTransferPreReq) (resp *types.C2CTransferPreRsp, err error) {
	itgRsp, err := l.svcCtx.TradeItgRpcClient.C2CTransferPre(l.ctx, &trade_itg_pb.C2CTransferPreReq{
		BuyerUserId: req.BuyerUserId,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.C2CTransferPreRsp{
		BuyerUserId:   req.BuyerUserId,
		TransactionId: itgRsp.TransactionId,
	}
	return
}
