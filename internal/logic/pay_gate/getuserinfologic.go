// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"

	"pay_gate/internal/svc"
	"pay_gate/internal/types"

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
	resp = &types.GetUserInfoRsp{
		UserId: req.UserId,
	}
	err = nil
	return
}
