package x

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type Config struct {
	Addr        string
	WebRoot     string
	LogLevel    slog.Level
	LogTimeZone string
}

// 앱 구조체
type App struct {
	Initialize      func()
	Finalize        func()
	OnShutdownErr   func(error)
	OnSignal        map[os.Signal]func()
	OnUnknownSignal func(os.Signal)
	Server          *http.Server
	Conf            *Config
	Logger          *slog.Logger
	Conns           map[string]*sql.DB
	Router          *Router
}

// 앱 생성자
func NewApp(config *Config) *App {
	if config.WebRoot == "" {
		config.WebRoot = "./www"
	}
	if err := os.MkdirAll(config.WebRoot, 0755); err != nil {
		panic(err)
	}
	loc, _ := time.LoadLocation(config.LogTimeZone)
	app := &App{
		Initialize:      func() {},
		Finalize:        func() {},
		OnShutdownErr:   func(err error) {},
		OnSignal:        make(map[os.Signal]func()),
		OnUnknownSignal: func(sig os.Signal) {},
		Conf:            config,
		Logger: slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: config.LogLevel,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						if t, ok := a.Value.Any().(time.Time); ok {
							return slog.Time(a.Key, t.In(loc))
						}
					}
					return a
				},
			},
		)),
		Conns:  map[string]*sql.DB{},
		Router: NewRouter(),
	}

	app.Server = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := NewContext(app, w, r)
			defer c.Recover()
			app.Router.ServeHTTP(c)
		}),
	}
	return app
}

func (a *App) CreateIndexFiles() {
	root := a.Conf.WebRoot
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		abs, _ := filepath.Abs(path)
		a.Logger.Debug("경로", "path", abs, "dir", d)
		if err != nil {
			return nil
		}
		if d.IsDir() {
			index := filepath.Join(path, "index.html")
			if _, err := os.Stat(index); err != nil {
				_ = os.WriteFile(index, []byte{}, 0644)
			}
		}
		return nil
	})
}

// 앱 실행
func (a *App) Run() {
	a.Server.Addr = a.Conf.Addr
	a.CreateIndexFiles()
	a.Initialize()

	a.Logger.Info("App initialized.")

	go func() {
		err := a.Server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	a.Wait()
}

// 시그널 콜백 등록
func (a *App) RegisterSignal(sig os.Signal, handler func()) {
	a.OnSignal[sig] = handler
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

	if err := a.Server.Shutdown(ctx); err != nil {
		a.OnShutdownErr(err)
	}

	// 서버가 정상적으로 내려간 뒤에 파이널 작업 실행
	a.Finalize()
	a.RemoveConns()
	a.Logger.Info("App finalized.")
}

func (a *App) RemoveConns() {
	for key, conn := range a.Conns {
		if conn != nil {
			if err := conn.Close(); err != nil {
				a.Logger.Warn("failed to close db connection", "key", key, "err", err)
			}
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
}

// 커넥션 가져오기
func (a *App) GetConn(key string) *sql.DB {
	return a.Conns[key]
}
