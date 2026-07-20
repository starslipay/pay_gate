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

type Bank2c_doLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBank2c_doLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Bank2c_doLogic {
	return &Bank2c_doLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Bank2c_doLogic) Bank2c_do(req *types.Bank2CDoReq) (resp *types.Bank2CDoRsp, err error) {
	bank2CDoRsp, err := l.svcCtx.TradeItgRpcClient.Bank2CDo(l.ctx, &trade_itg_pb.Bank2CDoReq{
		TransactionId: req.TransactionId,
		UserId:        req.UserId,
		BankType:      req.BankType,
		Amount:        req.Amount,
		Desc:          req.Desc,
	})
	if err != nil {
		return
	}
	resp = &types.Bank2CDoRsp{
		TransactionId: bank2CDoRsp.TransactionId,
		UserId:        bank2CDoRsp.UserId,
		IsRepeat:      bank2CDoRsp.IsRepeat,
	}
	return
}
