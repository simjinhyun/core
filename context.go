package x

import (
	"log"
	"net/http"
	"strings"
)

// Context : 요청/응답을 담는 컨텍스트
type Context struct {
	App      *App
	Req      *http.Request
	Res      http.ResponseWriter
	AppError *AppError
	Response struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
		Data    any    `json:"Data"`
		Elapsed int64  `json:"Elapsed"`
	}
}

func NewContext(a *App, w http.ResponseWriter, r *http.Request) *Context {
	return &Context{App: a, Req: r, Res: w}
}

func (c *Context) Reply() {
	// 응답 직렬화/쓰기 로직은 여기서 처리
}

// 로그 레벨 정의
const (
	LogDebug = iota
	LogInfo
	LogWarn
	LogError
)

// 문자열 → 레벨 매핑
func parseLogLevel(level string) int {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LogDebug
	case "INFO":
		return LogInfo
	case "WARN":
		return LogWarn
	case "ERROR":
		return LogError
	default:
		return LogInfo // 기본값
	}
}

// 현재 설정된 로그레벨 읽기
func (c *Context) currentLogLevel() int {
	if val, ok := c.App.Conf["LogLevel"].(string); ok {
		return parseLogLevel(val)
	}
	return LogInfo
}

// 내부 로그 출력 (필터링 적용)
func (c *Context) log(level int, msg string, args ...any) {
	if level < c.currentLogLevel() {
		return // 현재 설정보다 낮은 레벨은 무시
	}

	var prefix string
	switch level {
	case LogDebug:
		prefix = "[DEBUG] "
	case LogInfo:
		prefix = "[ INFO] "
	case LogWarn:
		prefix = "[ WARN] "
	case LogError:
		prefix = "[ERROR] "
	}

	if len(args) > 0 {
		log.Printf(prefix+msg, args...)
	} else {
		log.Print(prefix + msg)
	}
}

// 레벨별 헬퍼 메서드
func (c *Context) Debug(msg string, args ...any) { c.log(LogDebug, msg, args...) }
func (c *Context) Info(msg string, args ...any)  { c.log(LogInfo, msg, args...) }
func (c *Context) Warn(msg string, args ...any)  { c.log(LogWarn, msg, args...) }
func (c *Context) Error(msg string, args ...any) { c.log(LogError, msg, args...) }
