package util

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"time"
)

func GenMD5(input string) string {
	return fmt.Sprintf("%X", md5.Sum([]byte(input)))
}

func CheckUserToken(user_token, user_id, businessInfo string, tokenExpireTime int64) (isValid bool) {
	// 将token拆分为 md5Str 和 timestampStr
	md5Str := user_token[:32]
	timestampStr := user_token[32:]

	// 校验token中的签名是否正确
	calcMd5Str := GenMD5(businessInfo + user_id + timestampStr)
	if calcMd5Str != md5Str {
		return false
	}

	// 校验timestamp是否过期
	if timestampStr < strconv.FormatInt(time.Now().Unix()-tokenExpireTime, 10) {
		return false
	}

	return true
}
