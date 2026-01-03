package x

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

// 앱 구조체
type App struct {
	Server          *http.Server
	Initialize      func()
	Finalize        func()
	OnShutdownErr   func(error)
	OnSignal        map[os.Signal]func()
	OnUnknownSignal func(os.Signal)
	Conns           map[string]*sql.DB
	Router          *Router
	Logger          *slog.Logger
	Handler         *CustomHandler
}

// 앱 생성자
func NewApp(WebRoot string) *App {
	app := &App{
		Initialize:      func() {},
		Finalize:        func() {},
		OnShutdownErr:   func(err error) {},
		OnSignal:        make(map[os.Signal]func()),
		OnUnknownSignal: func(sig os.Signal) {},
		Conns:           map[string]*sql.DB{},
		Router:          NewRouter(WebRoot),
	}
	app.SetLogger(
		slog.LevelInfo,
		"",
		"2006.01.02 15:04:05 (MST)",
	)
	app.CreateIndexFiles(WebRoot)
	app.Server = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := NewContext(app, w, r)
			defer c.Recover()
			app.Router.ServeHTTP(c)
		}),
	}
	return app
}

func (a *App) SetLogger(l slog.Level, tz string, layout string) {
	a.Logger, a.Handler = NewLogger(l, tz, layout)
}

func (a *App) SetLevel(l slog.Level) {
	a.Handler.level.Set(l)
	a.Logger.Info("LogLevel changed", "Level", a.Handler.GetLevel())
}
func (a *App) CreateIndexFiles(WebRoot string) {
	filepath.WalkDir(WebRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}

		if d.IsDir() {
			index := filepath.Join(path, "index.html")
			_, statErr := os.Stat(index)

			if os.IsNotExist(statErr) {
				writeErr := os.WriteFile(index, []byte{}, 0644)
				if writeErr != nil {
					panic(writeErr)
				}
			} else if statErr != nil {
				panic(statErr)
			}
		}
		return nil
	})
}

// 앱 실행
func (a *App) Run(Addr string, shutdownTimeout time.Duration) {
	a.Server.Addr = Addr
	a.Initialize()

	a.Logger.Info("Logger", "Level", a.Handler.GetLevel())
	a.Logger.Info("Logger", "Timezone", a.Handler.GetTimezone())
	a.Logger.Info("App initialized")

	go func() {
		a.Logger.Info("App listening", "Addr", Addr)
		err := a.Server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	a.Wait(shutdownTimeout)
}

// 시그널 콜백 등록
func (a *App) RegisterSignal(sig os.Signal, handler func()) {
	a.OnSignal[sig] = handler
}

// 시그널 처리
func (a *App) Wait(shutdownTimeout time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop)

	for {
		sig := <-stop

		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			a.Shutdown(shutdownTimeout)
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

func (a *App) Shutdown(shutdownTimeout time.Duration) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout*time.Second,
	)
	defer cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		a.OnShutdownErr(err)
	}

	// 서버가 정상적으로 내려간 뒤에 파이널 작업 실행
	a.Finalize()
	a.Logger.Info("App finalized")
	a.RemoveConns()
}

func (a *App) RemoveConns() {
	for key, conn := range a.Conns {
		if conn != nil {
			if err := conn.Close(); err != nil {
				a.Logger.Warn("failed to close db connection", "key", key, "err", err)
			}
			a.Logger.Info("Connection removed", "key", key)
		}
	}
}

func (a *App) AddConn(key, driver, dsn string) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	a.Conns[key] = db
	a.Logger.Info("Connection added", "key", key)
}

// 커넥션 가져오기
func (a *App) GetConn(key string) *sql.DB {
	return a.Conns[key]
}

// AppError 구조체
type AppError struct {
	Code string // 에러 코드 (예: "RecordNotFound", "ParameterRequired")
	Src  string
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
		Src:  fmt.Sprintf("%s:%d", filepath.Base(file), line),
		Err:  err,
		Data: data,
	}
}
