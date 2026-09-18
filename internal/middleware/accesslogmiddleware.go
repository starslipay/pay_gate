package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// 脱敏后展示的占位符
const maskPlaceholder = "***"

// SensitiveFields 需要脱敏的字段名集合(全小写匹配)
type SensitiveFields map[string]bool

// NewSensitiveFields 从字段名列表构建集合，全部转小写
func NewSensitiveFields(fields []string) SensitiveFields {
	m := make(SensitiveFields, len(fields))
	for _, f := range fields {
		m[strings.ToLower(f)] = true
	}
	return m
}

// IsSensitive 判断字段名是否需要脱敏
func (sf SensitiveFields) IsSensitive(field string) bool {
	return sf[strings.ToLower(field)]
}

// AccessLogMiddleware 访问日志中间件：记录请求和响应参数，并按字段名脱敏
// 签名兼容 go-zero rest.Middleware: func(http.HandlerFunc) http.HandlerFunc
func AccessLogMiddleware(sensitiveFields SensitiveFields) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 1. 读取请求 body 并放回(不消费)
			var reqBody []byte
			if r.Body != nil {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(reqBody))
			}

			// 2. 调用前打印入参(脱敏)
			maskedReq := maskJSON(reqBody, sensitiveFields)
			logx.WithContext(r.Context()).Infof(
				"[ACCESS-IN] %s %s\n=> req: %s",
				r.Method, r.URL.Path, maskedReq,
			)

			// 3. 包装 ResponseWriter 捕获响应体
			lrw := &logResponseWriter{ResponseWriter: w, buf: &bytes.Buffer{}}

			// 4. 执行下游 handler
			next(lrw, r)

			// 5. 调用后打印出参(脱敏)
			maskedResp := maskJSON(lrw.buf.Bytes(), sensitiveFields)
			logx.WithContext(r.Context()).Infof(
				"[ACCESS-OUT] %s %s | duration=%s\n<= resp: %s",
				r.Method, r.URL.Path, time.Since(start).String(), maskedResp,
			)
		}
	}
}

// logResponseWriter 包装 http.ResponseWriter，同时捕获写入的响应体
type logResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (w *logResponseWriter) Write(b []byte) (int, error) {
	w.buf.Write(b) // 先缓存一份
	return w.ResponseWriter.Write(b)
}

// maskJSON 对 JSON 字节数据进行脱敏处理
// 递归遍历 JSON，将 sensitiveFields 中的字段值替换为 ***
func maskJSON(data []byte, sf SensitiveFields) string {
	if len(data) == 0 {
		return ""
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		// 不是合法 JSON，直接返回原始字符串
		return string(data)
	}

	masked := maskValue(v, sf)
	out, err := json.Marshal(masked)
	if err != nil {
		return string(data)
	}
	return string(out)
}

// maskValue 递归脱敏
func maskValue(v interface{}, sf SensitiveFields) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, vv := range val {
			if sf.IsSensitive(k) {
				result[k] = maskPlaceholder
			} else {
				result[k] = maskValue(vv, sf)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = maskValue(item, sf)
		}
		return result
	default:
		return v
	}
}
