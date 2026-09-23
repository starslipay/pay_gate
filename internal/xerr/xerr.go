package xerr

import (
	"fmt"

	"github.com/starslipay/paycomm/xerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CodeMsg struct {
	Code int64
	Msg  string
}

func (c *CodeMsg) Error() string {
	return fmt.Sprintf("[%d]%s", c.Code, c.Msg)
}

var (
	ModuleId        = int64(455901)
	ModuleErrorBase = ModuleId * 1000
)

var (
	CodeErrUnknown        = ModuleErrorBase + 0
	CodeErrServerInternal = ModuleErrorBase + 1
	CodeErrCallRpc        = ModuleErrorBase + 2

	CodeErrParam                                   = ModuleErrorBase + 100
	CodeErrUserNotExist                            = ModuleErrorBase + 101
	CodeErrPasswordWrong                           = ModuleErrorBase + 102
	CodeErrUserAlreadyRegistered                   = ModuleErrorBase + 103
	CodeErrRelationStateNotRegisteringOrRegistered = ModuleErrorBase + 104
	CodeErrTokenMissing                            = ModuleErrorBase + 105
	CodeErrTokenInvalid                            = ModuleErrorBase + 106
	CodeErrUserIdMissing                           = ModuleErrorBase + 107
	CodeErrTooManyRequests                         = ModuleErrorBase + 108
)

var (
	ErrTokenInvalid    = NewError(CodeErrTokenInvalid, "token invalid")
	ErrUserIdMissing   = NewError(CodeErrUserIdMissing, "user id missing")
	ErrTokenMissing    = NewError(CodeErrTokenMissing, "token missing")
	ErrTooManyRequests = NewError(CodeErrTooManyRequests, "too many requests")
)

func NewError(code int64, msg string) *CodeMsg {
	return &CodeMsg{
		Code: code,
		Msg:  msg,
	}
}

func FromError(err error) *CodeMsg {
	if err == nil {
		return nil
	}
	if ce, ok := err.(*CodeMsg); ok {
		return ce
	}
	return NewError(CodeErrUnknown, "unknown error")
}

func ParseRPCError(err error) error {
	// 如果业务错误存在，传递业务错误
	bizError, isSuccessParse := xerror.ParseBizError(err)
	if isSuccessParse {
		return NewError(bizError.Code, bizError.Message)
	}

	// 如果不是业务错误，传递默认错误
	return NewError(CodeErrCallRpc, "RPC_ERROR:"+err.Error())
}

// HandleRPCError 处理RPC错误码
// err: rpc调用返回的错误
// serviceName: 服务名，用于非业务错误场景的错误信息
// 返回处理后的error，包含业务错误码或服务调用错误信息
func HandleRPCError(err error, serviceName string) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return NewError(1000, fmt.Sprintf("%v 服务调用失败: %v", serviceName, st.Message()))
	}

	switch st.Code() {
	case codes.Internal:
		if bizErr, ok := xerror.ParseBizError(err); ok {
			return NewError(bizErr.Code, bizErr.Message)
		}
		return NewError(1001, fmt.Sprintf("%v 服务内部错误", serviceName))
	case codes.Unavailable:
		return NewError(1002, fmt.Sprintf("%v 服务不可达", serviceName))
	case codes.DeadlineExceeded:
		return NewError(1003, fmt.Sprintf("%v 调用超时", serviceName))
	case codes.Canceled:
		return NewError(1004, fmt.Sprintf("%v 上下文已取消", serviceName))
	default:
		return NewError(1005, fmt.Sprintf("%v 服务调用失败: %v", serviceName, st.Message()))
	}
}
