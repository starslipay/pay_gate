package pay_gate

import (
	"net/http"

	"github.com/starslipay/pay_gate/internal/logic/pay_gate"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 退款
func RefundOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefundOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pay_gate.NewRefundOrderLogic(r.Context(), svcCtx)
		resp, err := l.RefundOrder(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
