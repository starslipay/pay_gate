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

type Get_user_flowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGet_user_flowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Get_user_flowLogic {
	return &Get_user_flowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Get_user_flowLogic) Get_user_flow(req *types.GetUserFlowReq) (resp *types.GetUserFlowRsp, err error) {
	// 优先读 Redis relation 缓存, 未命中或 Redis 不可用时降级 UserMgr 查 DB
	relationRsp, err := getRelationCached(l.ctx, l.svcCtx, req.UserId)
	if err != nil {
		return nil, err
	}

	userFlowRsp, err := l.svcCtx.AccountMgr.GetUserFlow(l.ctx, &account_mgr_pb.GetUserFlowReq{
		Uid:    relationRsp.Uid,
		Offset: req.Offset,
		Limit:  req.Limit,
	})

	if err != nil {
		return nil, xerr.HandleRPCError(err, "AccountMgr.GetUserFlow")
	}
	resp = &types.GetUserFlowRsp{
		UserId:     req.UserId,
		NextOffset: userFlowRsp.NextOffset,
		EndFlag:    userFlowRsp.EndFlag,
	}

	for _, userFlowTmp := range userFlowRsp.UserFlowList {
		userFlow := &types.UserFlow{
			TransactionId:      userFlowTmp.TransactionId,
			UserId:             userFlowTmp.UserId,
			CounterpartyUserId: userFlowTmp.CounterpartyUserId,
			InoutType:          userFlowTmp.InoutType,
			BizType:            userFlowTmp.BizType,
			Balance:            userFlowTmp.Balance,
			Amount:             userFlowTmp.Amount,
			Desc:               userFlowTmp.Desc,
			Memo:               userFlowTmp.Memo,
			CreateTime:         userFlowTmp.CreateTime,
		}
		resp.UserFlowList = append(resp.UserFlowList, *userFlow)
	}

	return
}
