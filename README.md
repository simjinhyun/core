# x Framework

가볍고 단순하지만 강력한 Go 웹 프레임워크 **x**  
핵심은 **일관성, 자유도, 보안성**입니다.

---

## 🚀 주요 특징

### 1. 에러 처리의 일관성
- 모든 예외는 **AppError**로 통일
- 성공(`OK`)과 실패(`RuntimeError`, `RecordNotFound` 등)를 동일한 구조로 응답
- 개발자는 로직만 작성, 응답 포맷은 프레임워크가 자동 처리

### 2. 심플하고 강력한 라우터
- **등록된 엔드포인트** → 핸들러 실행
- **등록되지 않은 요청** → `WebRoot`에서 정적 파일 서빙
- `CreateIndexFiles()`로 모든 디렉토리에 `index.html` 자동 생성 → 디렉토리 listing 보안 문제 차단

### 3. 리소스 관리 철학
- DB 연결은 앱 기동 시 강제 등록 (`AddConn`)
- 실패 시 즉시 종료 → 안정성 확보
- Initialize / Finalize 훅으로 다른 리소스도 자유롭게 관리 가능

### 4. 시그널 처리 자유도
- 기본 제공: `SIGINT`, `SIGTERM` → 우아한 종료
- 사용자 정의: `OnSignal`에 원하는 시그널 핸들러 등록
- 놓친 시그널도 `OnUnknownSignal`로 대응 → **기본은 안전, 확장은 자유**

---

## 🛡️ 보안성
- `http.ServeFile()`의 디렉토리 listing 문제를 원천 차단
- 모든 디렉토리에 최소한의 `index.html`을 자동 생성
- 운영자가 별도 설정하지 않아도 안전한 기본값 제공

---

## 🗝️ 철학
- **핵심은 강제**: 안정성과 보안은 프레임워크가 책임짐  
- **나머지는 자유**: 개발자는 원하는 대로 확장 가능  

---

## 📂 예시 코드

```go
// JSON 핸들러 등록
GetRouter().HandleJSON("/api/hello", func(c *Context) {
    c.JSON(map[string]string{"message": "Hello, x!"})
})

// HTML 핸들러 등록
GetRouter().HandleHTML("/hello", func(c *Context) {
    c.HTML("<h1>Hello, x!</h1>")
})
