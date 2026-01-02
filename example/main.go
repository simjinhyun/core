package main

import (
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/simjinhyun/x"
)

func main() {
	a := x.NewApp(&x.Config{
		Addr:        "localhost:7000",
		LogLevel:    slog.LevelDebug,
		LogTimeZone: "Asia/Seoul",
	})
	a.AddConn("1", "mysql", "root:Tldrmf#2013@tcp(10.0.0.200:3306)/testdb?timeout=5s&readTimeout=30s&writeTimeout=30s")
	a.Logger.Info("DB Connected")
	a.Run()
}
