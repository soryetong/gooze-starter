package gzmiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/gooze-starter/gooze"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/soryetong/gooze-starter/services/gzauth"
	"go.uber.org/zap"
)

type RequestLogData struct {
	Username   string `json:"username"`   // 用户名
	UserId     int64  `json:"userId"`     // 用户ID
	Method     string `json:"method"`     // 请求方法
	Path       string `json:"path"`       // 请求路径
	StatusCode int64  `json:"statusCode"` // 状态码
	Elapsed    string `json:"elapsed"`    // 耗时
	Msg        string `json:"msg"`        // 返回的msg
	Request    string `json:"request"`    // 请求参数
	Response   string `json:"response"`   // 返回参数
	Platform   string `json:"platform"`   // 平台
	Ip         string `json:"ip"`         // IP
	Address    string `json:"address"`    // 地址
}

func RequestLog() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path, id := gzutil.GetRequestPath(ctx.Request.URL.Path, "/api")
		body := make(map[string]interface{})
		if id != 0 {
			body["id"] = id
		}

		if ctx.Request.Body != nil {
			bodyPost, _ := io.ReadAll(ctx.Request.Body)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyPost))
			body["post"] = string(bodyPost)
		}

		query := ctx.Request.URL.RawQuery
		if query != "" {
			query, _ = url.QueryUnescape(query)
			for _, v := range strings.Split(query, "&") {
				kv := strings.Split(v, "=")
				if len(kv) == 2 {
					body[kv[0]] = kv[1]
				}
			}
		}

		request, _ := json.Marshal(body)
		userAgent := ctx.GetHeader("User-Agent")
		logData := RequestLogData{
			Method:   ctx.Request.Method,
			Path:     path,
			Request:  string(request),
			UserId:   gzauth.GetTokenValue[int64](ctx, "id"),
			Username: gzauth.GetTokenValue[string](ctx, "username"),
			Platform: gzutil.GetPlatform(userAgent) + " " + gzutil.GetBrowser(userAgent),
		}

		writer := &responseBodyWriter{
			ResponseWriter: ctx.Writer,
			body:           &bytes.Buffer{},
		}
		ctx.Writer = writer
		startTime := time.Now()

		ctx.Next()

		elapsedMs := time.Since(startTime).Seconds() * 1000
		logData.Elapsed = fmt.Sprintf("%.2f", elapsedMs)
		resp := &gooze.Response{}
		_ = json.Unmarshal([]byte(writer.body.String()), resp)
		logData.StatusCode = resp.Code
		logData.Msg = resp.Msg
		respData, _ := json.Marshal(resp.Data)
		logData.Response = string(respData)
		gooze.Log.Info("[RequestLog]请求响应日志", zap.Any("logData", logData))

		// 非 Get 请求把数据放入Context中
		if ctx.Request.Method != "GET" {
			ctx.Set("requestLogData", &logData)
		}
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseBodyWriter) WriteHeader(statusCode int) {
	if !r.Written() {
		r.ResponseWriter.WriteHeader(statusCode)
	}
}
