package x

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"maps"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
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

// 앱 구조체
type App struct {
	Initialize      func()
	Finalize        func()
	OnShutdownErr   func(error)
	OnSignal        map[os.Signal]func()
	OnUnknownSignal func(os.Signal)
	server          *http.Server
	ConfRaw         []byte
	Conf            map[string]any
}

var defaultAddr = ":8080"

// 앱 생성자
func NewApp() *App {
	app := &App{
		Initialize:      func() {},
		Finalize:        func() {},
		OnShutdownErr:   func(err error) {},
		OnSignal:        make(map[os.Signal]func()),
		OnUnknownSignal: func(sig os.Signal) {},
		Conf:            map[string]any{"Addr": defaultAddr},
	}

	helper := func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(app, w, r) // 이제 app을 전달 가능
		defer func() {
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
				ctx.AppError = appErr
			}
			ctx.Reply()
		}()
		GetRouter().ServeHTTP(ctx)
	}

	app.server = &http.Server{
		Handler: http.HandlerFunc(helper),
	}

	return app
}

// 설정 로딩
func (a *App) LoadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("config load error: %v, fallback to default", err)
		return
	}

	a.ConfRaw = data
	var values map[string]any
	if err := json.Unmarshal(a.ConfRaw, &values); err != nil {
		log.Printf("config parse error: %v, fallback to default", err)
		return
	}

	// 기본값 유지하면서 덮어쓰기
	maps.Copy(a.Conf, values)

	// Addr 키가 없으면 기본값 보장
	if _, ok := a.Conf["Addr"]; !ok {
		log.Printf("config missing Addr, fallback to default")
		a.Conf["Addr"] = defaultAddr
	}

	// WebRoot 키가 없으면 기본값 보장
	if _, ok := a.Conf["WebRoot"]; !ok {
		log.Printf("config missing WebRoot, fallback to current directory")
		a.Conf["WebRoot"] = "." // 현재 디렉토리를 기본 루트로
	}
}

func checkIndexFiles(root string) {
	missingCount := 0

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			index := filepath.Join(path, "index.html")
			if _, err := os.Stat(index); err != nil {
				log.Printf("warning: directory %s has no index.html", path)
				missingCount++
			}
		}
		return nil
	})

	if missingCount > 0 {
		log.Printf("total %d directories are missing index.html", missingCount)
	} else {
		log.Printf("all directories have index.html")
	}
}

// 앱 실행
func (a *App) Run() {
	//CLI 옵션 파싱
	configFile := flag.String("f", "config.json", "설정파일경로")
	flag.Parse()

	//서버실행
	a.LoadConfig(*configFile)
	a.server.Addr = a.Conf["Addr"].(string)
	// 웹루트 검사
	if root, ok := a.Conf["WebRoot"].(string); ok {
		checkIndexFiles(root)
	}
	a.Initialize()

	go func() {
		err := a.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("http server failed: ", err)
		}
	}()

	a.Wait()
}

// 시그널 콜백 등록
func (a *App) RegisterSignal(sig os.Signal, handler func()) {
	a.OnSignal[sig] = handler
}

func (a *App) HandleJSON(path string, h HandlerFunc) {
	GetRouter().HandleJSON(path, h)
}

func (a *App) HandleHTML(path string, h HandlerFunc) {
	GetRouter().HandleHTML(path, h)
}

// 시그널 처리
func (a *App) Wait() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop)

	for {
		sig := <-stop

		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			a.Shutdown()
			return
		default:
			if handler, ok := a.OnSignal[sig]; ok {
				handler()
			} else {
				a.OnUnknownSignal(sig)
			}
		}
	}
}

func (a *App) Shutdown() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.OnShutdownErr(err)
	}

	// 서버가 정상적으로 내려간 뒤에 파이널 작업 실행
	a.Finalize()
}
