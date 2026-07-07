// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"pay_gate/internal/logic/pay_gate"
	"pay_gate/internal/svc"
	"pay_gate/internal/types"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HealthReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := pay_gate.NewHealthLogic(r.Context(), svcCtx)
		resp, err := l.Health(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
