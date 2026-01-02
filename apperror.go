package x

import (
	"fmt"
	"runtime"
)

// AppError 구조체
type AppError struct {
	Code string         // 에러 코드 (예: "RecordNotFound", "ParameterRequired")
	File string         // 발생 파일
	Line int            // 발생 라인
	Err  error          // 원본 에러
	Data map[string]any // 메시지 조립용 데이터
}

// Panic 메서드
func (e *AppError) Panic() {
	panic(e)
}

// 헬퍼 함수: 에러 생성
func NewAppError(code string, err error, data map[string]any) *AppError {
	_, file, line, _ := runtime.Caller(1)
	return &AppError{
		Code: code,
		File: file,
		Line: line,
		Err:  err,
		Data: data,
	}
}

// String 메서드 (디버깅용)
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s:%d %v", e.Code, e.File, e.Line, e.Err)
}
