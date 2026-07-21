package response

import (
	"context"
	"net/http"

	"github.com/starslipay/pay_gate/internal/xerr"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func init() {
	// httpx.SetOkHandler: 设置全局成功响应包装器
	// 当调用httpx.OkJson/OkJsonCtx时，传入的v会先经过这个函数处理
	// 这里将业务数据v包装成统一格式{code, msg, data}
	httpx.SetOkHandler(func(ctx context.Context, v interface{}) interface{} {
		return Response{
			Code: 0,
			Msg:  "success",
			Data: v,
		}
	})

	// httpx.SetErrorHandler: 设置全局错误响应处理函数
	// 当调用httpx.Error时，err会经过这个函数处理，返回(HTTP状态码, 响应体)
	// 这里将错误解析为统一格式{code, msg, data}
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		logx.WithContext(context.Background()).Errorf("http error: %v", err)
		ce := xerr.FromError(err)
		return http.StatusOK, Response{
			Code: ce.Code,
			Msg:  ce.Msg,
			Data: nil,
		}
	})

	// httpx.SetErrorHandlerCtx: 设置带上下文的全局错误响应处理函数
	// 当调用httpx.ErrorCtx时，err会经过这个函数处理，支持日志追踪
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, interface{}) {
		logx.WithContext(ctx).Errorf("http error ctx: %v", err)
		ce := xerr.FromError(err)
		return http.StatusOK, Response{
			Code: ce.Code,
			Msg:  ce.Msg,
			Data: nil,
		}
	})
}
