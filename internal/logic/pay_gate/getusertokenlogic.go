// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/pay_gate/internal/xerr"
	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/user_mgr/user_mgr_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type Get_user_tokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGet_user_tokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Get_user_tokenLogic {
	return &Get_user_tokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Get_user_tokenLogic) Get_user_token(req *types.GetUserTokenReq) (resp *types.GetUserTokenRsp, err error) {
	userMgrResp, err := l.svcCtx.UserMgrRpcClient.GetUserToken(l.ctx, &user_mgr_pb.GetUserTokenReq{
		UserId:       req.UserId,
		Password:     req.Password,
		BusinessInfo: req.BusinessInfo,
	})
	if err != nil {
		bizError, isSuccessParse := xerror.ParseBizError(err)
		if isSuccessParse {
			return nil, xerr.NewError(bizError.Code, bizError.Message)
		}

		return nil, xerr.NewError(xerr.CodeErrUnknown, "UNKNOWN:"+err.Error())
	}
	resp = &types.GetUserTokenRsp{
		UserId:    userMgrResp.UserId,
		UserToken: userMgrResp.UserToken,
	}
	return
}
