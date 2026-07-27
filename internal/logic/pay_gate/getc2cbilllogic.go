// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/paycomm/xerror"

	"github.com/zeromicro/go-zero/core/logx"
)

type Get_c2c_billLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGet_c2c_billLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Get_c2c_billLogic {
	return &Get_c2c_billLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Get_c2c_billLogic) Get_c2c_bill(req *types.GetC2CBillReq) (resp *types.GetC2CBillRsp, err error) {
	qryResp, err := l.svcCtx.AccountMgr.GetC2CBill(l.ctx, &account_mgr_pb.GetC2CBillReq{
		TransactionId: req.TransactionId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "AccountMgr.GetC2CBill")
	}

	resp = &types.GetC2CBillRsp{
		TransactionId: qryResp.TransactionId,
		BuyerUserId:   qryResp.BuyerUserId,
		SellerUserId:  qryResp.SellerUserId,
		Amount:        qryResp.Amount,
		Desc:          qryResp.Desc,
		PayTime:       qryResp.PayTime,
	}
	return
}
