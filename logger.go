package x

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type CustomHandler struct {
	level  *slog.LevelVar
	writer *os.File
	loc    *time.Location
	layout string
	attrs  []slog.Attr
	group  string
}

func NewCustomHandler(l *slog.LevelVar, loc *time.Location, layout string) *CustomHandler {
	return &CustomHandler{
		level:  l,
		writer: os.Stdout,
		loc:    loc,
		layout: layout,
	}
}

func (h *CustomHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level.Level()
}

func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	ts := r.Time.In(h.loc).Format(h.layout)

	// 소스 파일/라인 추출
	src := ""
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		src = fmt.Sprintf("(%s:%d)", filepath.Base(f.File), f.Line)
	}

	// 공통 헤더 출력
	fmt.Fprintf(h.writer, "%s %-5s %s %s", ts, r.Level.String(), src, r.Message)

	// Attrs를 JSON으로 직렬화
	attrs := make(map[string]any)
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	if len(attrs) > 0 {
		b, _ := json.Marshal(attrs)
		fmt.Fprintf(h.writer, " %s", b)
	}

	fmt.Fprintln(h.writer)
	return nil
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// 새로운 핸들러 복제해서 attrs 추가
	newH := *h
	newH.attrs = append(newH.attrs, attrs...)
	return &newH
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	// 그룹 이름은 단순히 prefix로만 저장
	newH := *h
	newH.group = name
	return &newH
}

func (h *CustomHandler) GetLevel() slog.Level { return h.level.Level() }
func (h *CustomHandler) GetTimezone() string  { return h.loc.String() }

func NewLogger(
	l slog.Level, tz string, layout string,
) (*slog.Logger, *CustomHandler) {
	lv := new(slog.LevelVar)
	lv.Set(l)

	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.Local
	}

	handler := NewCustomHandler(lv, loc, layout)
	return slog.New(handler), handler
}
