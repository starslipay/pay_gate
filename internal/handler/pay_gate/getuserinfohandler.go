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

// get_user_info
func Get_user_infoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUserInfoReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := pay_gate.NewGet_user_infoLogic(r.Context(), svcCtx)
		resp, err := l.Get_user_info(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
