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

type C2c_transfer_doLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewC2c_transfer_doLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2c_transfer_doLogic {
	return &C2c_transfer_doLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *C2c_transfer_doLogic) C2c_transfer_do(req *types.C2CTransferDoReq) (resp *types.C2CTransferDoRsp, err error) {
	c2CDoRsp, err := l.svcCtx.TradeItgRpcClient.C2CTransferDo(l.ctx, &trade_itg_pb.C2CTransferDoReq{
		TransactionId: req.TransactionId,
		BuyerUserId:   req.BuyerUserId,
		SellerUserId:  req.SellerUserId,
		Amount:        req.Amount,
		VerifyType:    req.VerifyType,
		Password:      req.Password,
	})
	if err != nil {
		return
	}
	resp = &types.C2CTransferDoRsp{
		TransactionId: c2CDoRsp.TransactionId,
		BuyerUserId:   c2CDoRsp.BuyerUserId,
		SellerUserId:  c2CDoRsp.SellerUserId,
		IsRepeat:      c2CDoRsp.IsRepeat,
	}
	return
}
