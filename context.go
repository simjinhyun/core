package x

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
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
	a.Logger.Debug("[REQ]"+c.ReqID, "Method", c.Req.Method, "Path", c.Req.URL.Path)
	return c
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
		default:
			appErr = NewAppError("RuntimeError", fmt.Errorf("%v", rec), nil)
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
		// 정적 파일 서빙은 응답생략
	}
	c.App.Logger.Debug("[RES]"+c.ReqID, "Code", c.Response.Code)
}

func (c *Context) ReplyJSON() {
	c.Res.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(c.Res).Encode(c.Response); err != nil {
		// 여기서는 panic 걸면 안 되고 안전하게 fallback
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
