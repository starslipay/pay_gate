// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/pay_gate/internal/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type Query_c2b_billLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQuery_c2b_billLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Query_c2b_billLogic {
	return &Query_c2b_billLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Query_c2b_billLogic) Query_c2b_bill(req *types.QueryC2BBillReq) (resp *types.QueryC2BBillRsp, err error) {
	qryResp, err := l.svcCtx.AccountMgr.QueryC2BBill(l.ctx, &account_mgr_pb.QueryC2BBillReq{
		TransactionId: req.TransactionId,
	})
	if err != nil {
		return nil, xerr.HandleRPCError(err, "AccountMgr.QueryC2BBill")
	}
	// 返回 account_mgr 查询到的全部单据信息
	resp = &types.QueryC2BBillRsp{
		TransactionId: qryResp.TransactionId,
		OutTradeNo:    qryResp.OutTradeNo,
		Uid:           qryResp.Uid,
		UserId:        qryResp.UserId,
		MerchantUid:   qryResp.MerchantUid,
		MerchantId:    qryResp.MerchantId,
		Amount:        qryResp.Amount,
		State:         qryResp.State,
		PayTime:       qryResp.PayTime,
		DeductToken:   qryResp.DeductToken,
	}
	return
}
