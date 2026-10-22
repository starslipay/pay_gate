// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/pay_gate/internal/xerr"
	"github.com/starslipay/user_mgr/user_mgr_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type Cancel_userLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancel_userLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Cancel_userLogic {
	return &Cancel_userLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Cancel_userLogic) Cancel_user(req *types.CancelUserReq) (resp *types.CancelUserRsp, err error) {
	// 调用user_mgr服务注销用户(软删除, 注销后禁止重新注册)
	cancelUserRsp, err := l.svcCtx.UserMgr.CancelUser(l.ctx, &user_mgr_pb.CancelUserReq{
		UserId: req.UserId,
	})
	if err != nil {
		return nil, xerr.HandleRPCError(err, "UserMgr.CancelUser")
	}
	resp = &types.CancelUserRsp{
		UserId: cancelUserRsp.UserId,
	}
	err = nil
	return
}
