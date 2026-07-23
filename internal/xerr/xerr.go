package xerr

import "fmt"

type CodeMsg struct {
	Code int64
	Msg  string
}

func (c *CodeMsg) Error() string {
	return fmt.Sprintf("[%d]%s", c.Code, c.Msg)
}

var (
	ModuleId        = int64(10000)
	ModuleErrorBase = ModuleId * 10000
)

// FromError 转换为CodeMsg
func FromError(err error) *CodeMsg {
	if err == nil {
		return nil
	}
	if ce, ok := err.(*CodeMsg); ok {
		return ce
	}
	return NewError(CodeErrUnknown, "unknown error")
}

// IsErrCode 判断错误是否为指定的错误码
func IsErrCode(err error, code int64) bool {
	if ce := FromError(err); ce != nil {
		return ce.Code == code
	}
	return false
}

var (
	CodeErrUnknown        = ModuleErrorBase + 0
	CodeErrServerInternal = ModuleErrorBase + 1

	CodeErrParam                                   = ModuleErrorBase + 1000
	CodeErrUserNotExist                            = ModuleErrorBase + 1001
	CodeErrPasswordWrong                           = ModuleErrorBase + 1002
	CodeErrUserAlreadyRegistered                   = ModuleErrorBase + 1003
	CodeErrRelationStateNotRegisteringOrRegistered = ModuleErrorBase + 1004
	CodeErrTokenMissing                            = ModuleErrorBase + 1005
	CodeErrTokenInvalid                            = ModuleErrorBase + 1006
	CodeErrUserIdMissing                           = ModuleErrorBase + 1007
)

var (
	ErrTokenInvalid  = NewError(CodeErrTokenInvalid, "token invalid")
	ErrUserIdMissing = NewError(CodeErrUserIdMissing, "user id missing")
	ErrTokenMissing  = NewError(CodeErrTokenMissing, "token missing")
)

func NewError(code int64, msg string) *CodeMsg {
	return &CodeMsg{
		Code: code,
		Msg:  msg,
	}
}
