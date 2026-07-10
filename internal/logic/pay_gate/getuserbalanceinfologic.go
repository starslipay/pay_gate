// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/user_mgr/user_mgr_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	QryModeSlave  = 1 // 从库查询
	QryModeMaster = 2 // 主库查询
)

type Get_user_balance_infoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGet_user_balance_infoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Get_user_balance_infoLogic {
	return &Get_user_balance_infoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Get_user_balance_infoLogic) Get_user_balance_info(req *types.GetUserBalanceInfoReq) (resp *types.GetUserBalanceInfoRsp, err error) {
	relationRsp, err := l.svcCtx.UserMgrRpcClient.GetRelation(l.ctx, &user_mgr_pb.GetRelationReq{
		UserId: req.UserId,
	})
	if err != nil {
		l.Logger.Errorf("GetRelation failed, err: %v", err)
		return
	}

	getUserBalanceInfoRsp, err := l.svcCtx.AccountMgrRpcClient.GetUserBalanceInfo(l.ctx, &account_mgr_pb.GetUserBalanceInfoReq{
		Uid:     relationRsp.Uid,
		QryMode: QryModeSlave, // 查询从库
	})
	if err != nil {
		l.Logger.Errorf("GetUserBalanceInfo failed, err: %v", err)
		return
	}

	return &types.GetUserBalanceInfoRsp{
		UserId:  req.UserId,
		Balance: getUserBalanceInfoRsp.Balance,
		CurType: getUserBalanceInfoRsp.CurType,
	}, nil
}
