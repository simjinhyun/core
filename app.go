package x

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 앱 구조체
type App struct {
	Initialize     func()
	RequestHandler func(*Context)
	Finalize       func()
	OnShutdownErr  func(error)
	server         *http.Server
}

// 앱 생성자
func NewApp(
	initialize func(),
	requestHandler func(*Context),
	finalize func(),
	onShutdownErr func(error),
) *App {
	if initialize == nil {
		initialize = func() {}
	}
	if finalize == nil {
		finalize = func() {}
	}
	if onShutdownErr == nil {
		onShutdownErr = func(err error) {}
	}

	helper := func(w http.ResponseWriter, r *http.Request) {
		requestHandler(NewContext(w, r))
	}

	return &App{
		Initialize:     initialize,
		RequestHandler: requestHandler,
		Finalize:       finalize,
		OnShutdownErr:  onShutdownErr,
		server: &http.Server{
			Handler: http.HandlerFunc(helper),
		},
	}
}

const Pipe = "x.pipe"

// ----------------------------------------------------
// Run : 서버 실행 메인 함수
// ----------------------------------------------------
func (a *App) Run() {
	//CLI 옵션 파싱
	configFile := flag.String("f", "config.json", "설정파일경로")
	T := flag.Bool("T", false, "설정파일보기")
	flag.Parse()

	//실제 로딩된 설정파일내용 응답
	if *T {
		ShowConfig()
		return
	}

	//서버실행
	cx := LoadConfig(*configFile)

	Debug("서버리스닝")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 필요할 때 cancel() 호출하면 ReadPipe 고루틴 종료

	err := ReadPipe(ctx, Pipe, func(data []byte) {
		Debug(fmt.Sprintf("ReadPipe %s", data))
		WritePipe(string(data), []byte(cx.Data+"\n"))
		WritePipe(string(data), []byte("EXIT\n"))
	})
	if err != nil {
		Debug("ReadPipe 에러: " + err.Error())
	}

	Debug("앱초기화")
	// --- 서버 모드 (기본 실행) ---
	a.server.Addr = cx.Values["Addr"].(string)
	a.Initialize()

	// HTTP 서버 실행
	go func() {

		err := a.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("http server failed: %w", err))
		}
	}()

	a.Wait()
}

func (a *App) Wait() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		sig := <-stop

		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			a.Finalize()

			_ = os.Remove(Pipe)

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			if err := a.server.Shutdown(ctx); err != nil {
				a.OnShutdownErr(err)
			}
			return

		case syscall.SIGUSR1:
			fmt.Println("Received SIGUSR1.")

		default:
			fmt.Println("Ignoring signal:", sig)
		}
	}
}
