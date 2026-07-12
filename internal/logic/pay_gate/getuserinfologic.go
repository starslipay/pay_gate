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

type Get_user_infoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// get_user_info
func NewGet_user_infoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Get_user_infoLogic {
	return &Get_user_infoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Get_user_infoLogic) Get_user_info(req *types.GetUserInfoReq) (resp *types.GetUserInfoRsp, err error) {
	// 调用user_mgr服务
	userInfo, err := l.svcCtx.UserMgrRpcClient.GetUserInfo(l.ctx, &user_mgr_pb.GetUserInfoReq{
		UserId: req.UserId,
	})
	if err != nil {
		return nil, xerr.NewServerInternalError("GetUserInfo failed:" + err.Error())
	}
	resp = &types.GetUserInfoRsp{
		UserId:  userInfo.UserId,
		Name:    userInfo.Name,
		Gender:  userInfo.Gender,
		Age:     userInfo.Age,
		Address: userInfo.Address,
		Phone:   userInfo.Phone,
		Email:   userInfo.Email,
		IdType:  userInfo.IdType,
		IdCard:  userInfo.IdCard,
	}
	err = nil
	return
}
