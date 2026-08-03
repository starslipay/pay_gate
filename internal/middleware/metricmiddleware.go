package middleware

import (
	"context"
	"net/http"
)

// ctxKey 私有类型, 避免 ctx key 冲突
type ctxKey struct{}

var methodKey = ctxKey{}

// MetricMethodMiddleware 全局中间件: 把请求路径写入 ctx,
// 供 response 层(SetOkHandler/SetErrorHandlerCtx)作为 method 标签读取。
// 因为 go-zero 的响应处理函数签名只有 ctx, 拿不到 *http.Request, 故此中转。
func MetricMethodMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), methodKey, r.URL.Path)
		next(w, r.WithContext(ctx))
	}
}

// MethodFromCtx 从 ctx 取回请求路径, 取不到返回 "unknown"
func MethodFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(methodKey).(string); ok && v != "" {
		return v
	}
	return "unknown"
}
