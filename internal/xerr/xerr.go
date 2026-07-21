package xerr

import "fmt"

type CodeMsg struct {
	Code int
	Msg  string
}

func (c *CodeMsg) Error() string {
	return fmt.Sprintf("[%d]%s", c.Code, c.Msg)
}

var (
	ModuleId        = 10000
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
	return ErrUnknown
}

// IsErrCode 判断错误是否为指定的错误码
func IsErrCode(err error, code int) bool {
	if ce := FromError(err); ce != nil {
		return ce.Code == code
	}
	return false
}

var (
	ErrServerInternal = newError(ModuleErrorBase+0, "server internal error")
	ErrUnknown        = newError(ModuleErrorBase+1, "unknown error")

	ErrParam                                   = newError(ModuleErrorBase+1000, "param error")
	ErrUserNotExist                            = newError(ModuleErrorBase+1001, "user not exist")
	ErrPasswordWrong                           = newError(ModuleErrorBase+1002, "password wrong")
	ErrUserAlreadyRegistered                   = newError(ModuleErrorBase+1003, "user already registered")
	ErrRelationStateNotRegisteringOrRegistered = newError(ModuleErrorBase+1004, "relation state is not registering or registered")
	ErrTokenMissing                            = newError(ModuleErrorBase+1005, "user_token is missing")
	ErrTokenInvalid                            = newError(ModuleErrorBase+1006, "user_token is invalid")
	ErrUserIdMissing                           = newError(ModuleErrorBase+1007, "user_id is missing")
)

func newError(code int, msg string) *CodeMsg {
	return &CodeMsg{
		Code: code,
		Msg:  msg,
	}
}

func NewParamError(msg string) *CodeMsg {
	return newError(ErrParam.Code, msg)
}

func NewServerInternalError(msg string) *CodeMsg {
	return newError(ErrServerInternal.Code, msg)
}
