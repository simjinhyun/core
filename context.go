package x

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/simjinhyun/x/util"
)

// Context : 요청/응답을 담는 컨텍스트
type Context struct {
	App       *App
	Req       *http.Request
	Res       http.ResponseWriter
	AppError  *AppError
	RouteType string
	Store     map[string]any //핸들러 체인들이 자유롭게 데이터 담을 수 있게
	ReqID     string
	ReqTime   time.Time
	ReqBody   string
	RemoteIP  string
	Response  struct {
		Code    string
		Message string
		Data    any
		Elapsed string
	}
}

func NewContext(a *App, w http.ResponseWriter, r *http.Request) *Context {
	now := time.Now()
	c := &Context{
		App:      a,
		Req:      r,
		Res:      w,
		Store:    map[string]any{},
		ReqID:    util.EncodeToBase62(uint64(now.UnixNano())),
		ReqTime:  now,
		RemoteIP: getClientIP(r),
	}
	c.CopyBody()
	return c
}

func (c *Context) CopyBody() {
	if c.Req.Body == nil {
		c.ReqBody = ""
		return
	}

	ct := c.Req.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/") {
		c.ReqBody = ""
		return
	}

	// 최대 1MB까지만 읽기
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Req.Body, 1024*1024))
	if err != nil {
		c.ReqBody = ""
		return
	}

	// Body 복원
	c.Req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	c.ReqBody = string(bodyBytes)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ipStr := strings.TrimSpace(ips[0])
			if net.ParseIP(ipStr) != nil {
				return ipStr
			}
		}
	}

	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		if net.ParseIP(xRealIP) != nil {
			return xRealIP
		}
	}

	if cfConnectingIP := r.Header.Get("CF-Connecting-IP"); cfConnectingIP != "" {
		if net.ParseIP(cfConnectingIP) != nil {
			return cfConnectingIP
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (c *Context) Recover() {
	if rec := recover(); rec != nil {
		var appErr *AppError
		switch e := rec.(type) {
		case *AppError:
			appErr = e
		case error:
			appErr = NewAppError("RuntimeError", e, nil)
			c.App.Logger.Error(fmt.Sprintf("%s", debug.Stack()))
		default:
			appErr = NewAppError("RuntimeError", fmt.Errorf("%v", rec), nil)
			c.App.Logger.Error(fmt.Sprintf("%s", debug.Stack()))
		}
		c.AppError = appErr
	} else {
		c.AppError = NewAppError("OK", nil, nil)
	}
	c.Reply()
}

func (c *Context) Reply() {
	c.Response.Code = c.AppError.Code
	c.Response.Elapsed = time.Since(c.ReqTime).String()
	switch c.RouteType {
	case "JSON":
		c.ReplyJSON()
	case "HTML":
		c.ReplyHTML()
	default:
		// 정적 파일 서빙은 http.ServeFile 가 직접 응답함
	}

	//디버그 로그 (운영 성능 영향 제로)
	c.App.Logger.Debug(
		c.PrependReqID("[RES]"),
		c.Req.Method, c.Req.URL.Path,
		"AppError", c.AppError,
		"Elapsed", c.Response.Elapsed,
	)
}

func (c *Context) PrependReqID(msg string) string {
	return fmt.Sprintf("%s %s", c.ReqID, msg)
}

func (c *Context) ReplyJSON() {
	c.Res.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(c.Res).Encode(c.Response); err != nil {
		http.Error(c.Res, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) ReplyHTML() {
	c.Res.Header().Set("Content-Type", "text/html; charset=utf-8")

	if c.Response.Code == "OK" {
		if html, ok := c.Response.Data.(string); ok {
			fmt.Fprint(c.Res, html)
		} else {
			c.ReplyJSON()
		}
	} else {
		fmt.Fprintf(
			c.Res,
			"<html><body><h1>Error: %s</h1></body></html>",
			c.Response.Code,
		)
	}
}

// 값 저장
func (c *Context) Set(key string, value any) {
	c.Store[key] = value
}

// 범용 Get: 그냥 any 반환
func (c *Context) Get(key string) any {
	return c.Store[key]
}

func (c *Context) GetFloat64(key string) float64 {
	if v, ok := c.Store[key].(float64); ok {
		return v
	}
	return 0
}

func (c *Context) GetFloat32(key string) float32 {
	if v, ok := c.Store[key].(float32); ok {
		return v
	}
	return 0
}

func (c *Context) GetInt64(key string) int64 {
	if v, ok := c.Store[key].(int64); ok {
		return v
	}
	return 0
}

func (c *Context) GetInt32(key string) int32 {
	if v, ok := c.Store[key].(int32); ok {
		return v
	}
	return 0
}

func (c *Context) GetUint(key string) uint {
	if v, ok := c.Store[key].(uint); ok {
		return v
	}
	return 0
}

func (c *Context) GetBytes(key string) []byte {
	if v, ok := c.Store[key].([]byte); ok {
		return v
	}
	return nil
}

func (c *Context) GetRune(key string) rune {
	if v, ok := c.Store[key].(rune); ok {
		return v
	}
	return 0
}

func (c *Context) GetError(key string) error {
	if v, ok := c.Store[key].(error); ok {
		return v
	}
	return nil
}
