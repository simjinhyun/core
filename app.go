package x

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	Initialize     func() error
	RequestHandler func(http.ResponseWriter, *http.Request)
	Finalize       func() error
	server         *http.Server
}

func New(addr string, init func() error, handler func(http.ResponseWriter, *http.Request), finalize func() error) *App {
	return &App{
		Initialize:     init,
		RequestHandler: handler,
		Finalize:       finalize,
		server: &http.Server{
			Addr:    addr,
			Handler: http.HandlerFunc(handler),
		},
	}
}

func (a *App) Run() {
	// 1. 초기화 콜백 실행
	if a.Initialize != nil {
		if err := a.Initialize(); err != nil {
			panic(err)
		}
	}

	// 2. 서버 실행
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// 3. 종료 시그널 대기 후 Finalize 콜백 실행
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if a.Finalize != nil {
		_ = a.Finalize()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = a.server.Shutdown(ctx)
}
