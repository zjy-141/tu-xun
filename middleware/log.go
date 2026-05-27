package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
	"tu-xun/config"
	"tu-xun/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// formatBody 安全地格式化 body 字节内容用于日志。
// 对于文本内容返回字符串，对于二进制内容返回类型和长度摘要。
func formatBody(body []byte, contentType string) string {
	if len(body) == 0 {
		return ""
	}
	if isBinaryContent(body, contentType) {
		if contentType == "" {
			contentType = "unknown"
		}
		return fmt.Sprintf("[binary data, type: %s, length: %d bytes]", contentType, len(body))
	}
	return string(body)
}

// isBinaryContent 根据 Content-Type 和实际字节检测是否为二进制内容
func isBinaryContent(data []byte, contentType string) bool {
	if contentType != "" {
		ct := strings.ToLower(contentType)
		// 文本类 Content-Type，直接判定为非二进制
		if strings.HasPrefix(ct, "text/") ||
			strings.Contains(ct, "json") ||
			strings.Contains(ct, "xml") ||
			strings.Contains(ct, "javascript") ||
			strings.Contains(ct, "x-www-form-urlencoded") {
			return false
		}
		// 明确的二进制 Content-Type
		if strings.HasPrefix(ct, "image/") ||
			strings.HasPrefix(ct, "audio/") ||
			strings.HasPrefix(ct, "video/") ||
			strings.HasPrefix(ct, "application/octet-stream") ||
			strings.HasPrefix(ct, "application/pdf") ||
			strings.HasPrefix(ct, "application/zip") ||
			strings.HasPrefix(ct, "application/gzip") ||
			strings.HasPrefix(ct, "application/x-protobuf") ||
			strings.HasPrefix(ct, "application/grpc") {
			return true
		}
	}
	// 兜底：检查实际字节中是否包含 \x00（空字节是二进制数据的强特征）
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		w := &logger.ResponseBodyWriter{Body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		// Read and store the request body
		var requestBody []byte
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = make([]byte, len(bodyBytes))
				copy(requestBody, bodyBytes)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset request body
			}
		}

		// Create a copy of the context for logging
		// logContext := *c
		logContext := struct {
			Writer   gin.ResponseWriter
			Request  *http.Request
			ClientIP func() string
		}{
			Writer:   c.Writer,
			Request:  c.Request,
			ClientIP: c.ClientIP,
		}

		c.Next()
		go func() {
			select {
			case <-config.SkipSignalChan:
				return
			default:
				goto log
			}
		log:
			status := logContext.Writer.Status()
			path := logContext.Request.URL.Path
			query := logContext.Request.URL.RawQuery
			cost := time.Since(start)
			method := logContext.Request.Method
			clientIP := logContext.ClientIP()
			userAgent := logContext.Request.UserAgent()

			if logger.GinLogger.Level == logrus.DebugLevel {
				if logContext.Writer != nil {
					responseHeaders := logContext.Writer.Header()
					responseBody := w.Body.Bytes()
					requestHeaders, _ := httputil.DumpRequest(logContext.Request, false)

					// 获取请求和响应的 Content-Type，用于判断是否为二进制内容
					reqContentType := logContext.Request.Header.Get("Content-Type")
					respContentType := responseHeaders.Get("Content-Type")

					logger.GinLogger.WithFields(logrus.Fields{
						"\nmethod":           method,
						"\nurl":              path,
						"\nquery":            query,
						"\nclient_ip":        clientIP,
						"\nuser_agent":       userAgent,
						"\nstatus":           status,
						"\nduration":         cost,
						"\nrequest_headers":  string(requestHeaders),
						"\nrequest_body":     formatBody(requestBody, reqContentType),
						"\nresponse_headers": responseHeaders,
						"\nresponse_body":    formatBody(responseBody, respContentType),
					}).Debug()
				}
			} else {
				switch {
				case status >= http.StatusInternalServerError:
					logger.GinLogger.WithFields(logrus.Fields{
						"\nmethod":     method,
						"\nurl":        path,
						"\nquery":      query,
						"\nclient_ip":  clientIP,
						"\nuser_agent": userAgent,
						"\nStatus":     status,
						"\nduration":   cost}).Error()
				case status >= http.StatusBadRequest:
					logger.GinLogger.WithFields(logrus.Fields{
						"\nmethod":     method,
						"\nurl":        path,
						"\nquery":      query,
						"\nclient_ip":  clientIP,
						"\nuser_agent": userAgent,
						"\nstatus":     status,
						"\nduration":   cost}).Warn()
				default:
					logger.GinLogger.WithFields(logrus.Fields{
						"\nmethod":     method,
						"\nurl":        path,
						"\nquery":      query,
						"\nclient_ip:": clientIP,
						"\nuser_agent": userAgent,
						"\nstatus":     status,
						"\nduration":   cost}).Info()
				}
			}
		}()
	}
}

func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.GinLogger.Error("broken pipe: ", err, ". Request: ", string(httpRequest))
					c.Abort()
					return
				}

				//deeper stack
				pc, file, line, _ := runtime.Caller(3)
				func_name := runtime.FuncForPC(pc).Name()

				if stack {
					logger.GinLogger.Errorf("panic recovered: %v. Request: %s. File: %s, Line: %d, Function: %s. Stack: %s", err, string(httpRequest), file, line, func_name, string(debug.Stack()))
				} else {
					logger.GinLogger.Errorf("panic recovered: %v. Request: %s. File: %s, Line: %d, Function: %s", err, string(httpRequest), file, line, func_name)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
