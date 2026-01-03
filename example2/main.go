package main

import (
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/simjinhyun/x"
)

var (
	Version  string
	Revision string
	Date     string
)

func main() {
	a := x.NewApp("www")
	a.SetLogger(slog.LevelDebug, "Asia/Seoul", "2006.01.02 15:04:05 (MST)")
	a.Logger.Info("Build", "Version", Version)
	a.Logger.Info("Build", "Revision", Revision)
	a.Logger.Info("Build", "Date", Date)
	a.AddConn(
		"db1",
		"mysql",
		"root:Tldrmf#2013@tcp(10.0.0.200:3306)/testdb?timeout=5s&readTimeout=30s&writeTimeout=30s",
	)

	a.Router.HandleHTML("GET", "/hello", Hello)

	a.Run("localhost:7000", 5)
}

func Hello(c *x.Context) {
	c.Response.Data = "Hello World"
	// a := 0
	// _ = 1 / a
	panic("a")
	// x.NewAppError("XXX", nil, nil).Panic()
}
