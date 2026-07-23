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

type Update_user_infoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdate_user_infoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Update_user_infoLogic {
	return &Update_user_infoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Update_user_infoLogic) Update_user_info(req *types.UpdateUserInfoReq) (resp *types.UpdateUserInfoRsp, err error) {
	// 调用user_mgr服务
	UpdateUserInfoRsp, err := l.svcCtx.UserMgrRpcClient.UpdateUserInfo(l.ctx, &user_mgr_pb.UpdateUserInfoReq{
		UserId:  req.UserId,
		Name:    req.Name,
		Gender:  req.Gender,
		Age:     req.Age,
		Address: req.Address,
		Phone:   req.Phone,
		Email:   req.Email,
		IdType:  req.IdType,
		IdCard:  req.IdCard,
	})
	if err != nil {
		return nil, xerr.ParseRPCError(err)
	}
	resp = &types.UpdateUserInfoRsp{
		UserId: UpdateUserInfoRsp.UserId,
	}
	err = nil
	return
}
