// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/starslipay/user_mgr/user_mgr_pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var (
	ErrTokenMissing  = errors.New("user_token is missing")
	ErrTokenInvalid  = errors.New("user_token is invalid")
	ErrUserIdMissing = errors.New("user_id is missing")
)

type AuthInterceptorMiddleware struct {
	userMgrClient user_mgr_pb.UserMgrClient
}

func NewAuthInterceptorMiddleware(client user_mgr_pb.UserMgrClient) *AuthInterceptorMiddleware {
	return &AuthInterceptorMiddleware{
		userMgrClient: client,
	}
}

func (m *AuthInterceptorMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userId, userToken string

		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				logx.Error("parse request body error:", err)
				httpx.ErrorCtx(r.Context(), w, ErrTokenMissing)
				return
			}

			if token, ok := body["user_token"].(string); ok {
				userToken = token
			}
			if id, ok := body["user_id"].(string); ok {
				userId = id
			}
		} else {
			userToken = r.FormValue("user_token")
			userId = r.FormValue("user_id")
		}

		if userToken == "" {
			logx.Error("user_token is missing")
			httpx.ErrorCtx(r.Context(), w, ErrTokenMissing)
			return
		}

		if userId == "" {
			logx.Error("user_id is missing")
			httpx.ErrorCtx(r.Context(), w, ErrUserIdMissing)
			return
		}

		rsp, err := m.userMgrClient.CheckUserToken(r.Context(), &user_mgr_pb.CheckUserTokenReq{
			UserId:    userId,
			UserToken: userToken,
		})
		if err != nil {
			logx.Error("check user token rpc error:", err)
			httpx.ErrorCtx(r.Context(), w, ErrTokenInvalid)
			return
		}

		if rsp.GetValidStatus() != 1 {
			logx.Error("user_token is invalid, status:", rsp.GetValidStatus())
			httpx.ErrorCtx(r.Context(), w, ErrTokenInvalid)
			return
		}

		logx.Info("token validated successfully for user:", userId)
		next(w, r)
	}
}
