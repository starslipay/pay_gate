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

type C2bank_doLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewC2bank_doLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2bank_doLogic {
	return &C2bank_doLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *C2bank_doLogic) C2bank_do(req *types.C2BankDoReq) (resp *types.C2BankDoRsp, err error) {
	bank2CDoRsp, err := l.svcCtx.TradeItgRpcClient.C2BankDo(l.ctx, &trade_itg_pb.C2BankDoReq{
		TransactionId: req.TransactionId,
		UserId:        req.UserId,
		BankType:      req.BankType,
		Amount:        req.Amount,
		Desc:          req.Desc,
		VerifyType:    req.VerifyType,
		Password:      req.Password,
	})
	if err != nil {
		return nil, xerr.ParseRPCError(err)
	}
	resp = &types.C2BankDoRsp{
		TransactionId: bank2CDoRsp.TransactionId,
		UserId:        bank2CDoRsp.UserId,
		IsRepeat:      bank2CDoRsp.IsRepeat,
	}
	return
}
