package pay_gate

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/user_mgr/user_mgr_pb"

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
	// 1. 先用 merchant_id 查 merchant_uid 和 merchant_name
	merchantInfo, err := l.svcCtx.UserMgr.GetMerchantInfo(l.ctx, &user_mgr_pb.GetMerchantInfoReq{
		MerchantId: req.MerchantId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.GetMerchantInfo")
	}

	// 2. 再用 merchant_uid 查余额
	balanceResp, err := l.svcCtx.AccountMgr.GetMerchantBalanceInfo(l.ctx, &account_mgr_pb.GetMerchantBalanceInfoReq{
		MerchantUid: merchantInfo.MerchantUid,
		QryMode:     req.QryMode,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "AccountMgr.GetMerchantBalanceInfo")
	}

	return &types.GetMerchantBalanceInfoRsp{
		MerchantId:   merchantInfo.MerchantId,
		MerchantUid:  merchantInfo.MerchantUid,
		MerchantName: merchantInfo.MerchantName,
		Balance:      balanceResp.Balance,
		CurType:      balanceResp.CurType,
	}, nil
}
