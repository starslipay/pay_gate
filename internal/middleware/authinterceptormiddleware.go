// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/starslipay/pay_gate/internal/config"
	"github.com/starslipay/pay_gate/internal/util"
	"github.com/starslipay/pay_gate/internal/xerr"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthInterceptorMiddleware struct {
	config *config.Config
}

func NewAuthInterceptorMiddleware(config *config.Config) *AuthInterceptorMiddleware {
	return &AuthInterceptorMiddleware{
		config: config,
	}
}

func (m *AuthInterceptorMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userToken := r.Header.Get("UserToken")
		businessInfo := r.Header.Get("BusinessInfo")

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logx.Errorf("read request body error: %v", err)
			httpx.ErrorCtx(r.Context(), w, xerr.ErrTokenInvalid)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		var userId string
		if len(bodyBytes) > 0 {
			var reqBody struct {
				UserId string `json:"user_id"`
			}
			if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
				logx.Errorf("parse request body error: %v", err)
				httpx.ErrorCtx(r.Context(), w, xerr.ErrTokenInvalid)
				return
			}
			userId = reqBody.UserId
		}

		if userToken == "" || businessInfo == "" || userId == "" {
			logx.Errorf("missing auth params: userToken=%s, businessInfo=%s, userId=%s", userToken, businessInfo, userId)
			httpx.ErrorCtx(r.Context(), w, xerr.ErrTokenInvalid)
			return
		}

		if !util.CheckUserToken(userToken, userId, businessInfo, m.config.TokenExpireTime) {
			logx.Error("user_token validation failed")
			httpx.ErrorCtx(r.Context(), w, xerr.ErrTokenInvalid)
			return
		}

		logx.Infof("token validated successfully for user: %s", userId)
		next(w, r)
	}
}
