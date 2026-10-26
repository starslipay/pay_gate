package pay_gate

import (
	"net/http"

	"github.com/starslipay/pay_gate/internal/logic/pay_gate"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 查询充值提现单
func QueryRechargeWithdrawOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryRechargeWithdrawOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pay_gate.NewQueryRechargeWithdrawOrderLogic(r.Context(), svcCtx)
		resp, err := l.QueryRechargeWithdrawOrder(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
