// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/user_mgr/user_mgr_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type Reg_userLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReg_userLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Reg_userLogic {
	return &Reg_userLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Reg_userLogic) Reg_user(req *types.RegUserReq) (resp *types.RegUserRsp, err error) {
	// 调用user_mgr服务
	RegUserRsp, err := l.svcCtx.UserMgrRpcClient.RegUser(l.ctx, &user_mgr_pb.RegUserReq{
		UserId:   req.UserId,
		Password: req.Password,
		Name:     req.Name,
		Gender:   req.Gender,
		Age:      req.Age,
		Address:  req.Address,
		Phone:    req.Phone,
		Email:    req.Email,
		IdType:   req.IdType,
		IdCard:   req.IdCard,
	})
	if err != nil {
		l.Logger.Errorf("RegUser failed, err: %v", err)
		return
	}
	resp = &types.RegUserRsp{
		UserId: RegUserRsp.UserId,
	}
	err = nil
	return
}
