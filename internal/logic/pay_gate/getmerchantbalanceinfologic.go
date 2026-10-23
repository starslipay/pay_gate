package pay_gate

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/paycomm/xerror"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMerchantBalanceInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询商户余额
func NewGetMerchantBalanceInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMerchantBalanceInfoLogic {
	return &GetMerchantBalanceInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMerchantBalanceInfoLogic) GetMerchantBalanceInfo(req *types.GetMerchantBalanceInfoReq) (resp *types.GetMerchantBalanceInfoRsp, err error) {
	rpcResp, err := l.svcCtx.AccountMgr.GetMerchantBalanceInfo(l.ctx, &account_mgr_pb.GetMerchantBalanceInfoReq{
		MerchantUid: req.MerchantUid,
		QryMode:     req.QryMode,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "查询商户余额")
	}

	return &types.GetMerchantBalanceInfoRsp{
		MerchantUid: rpcResp.MerchantUid,
		MerchantId:  rpcResp.MerchantId,
		Balance:     rpcResp.Balance,
		CurType:     rpcResp.CurType,
	}, nil
}
