package main

import (
	"github.com/simjinhyun/x"
)

func main() {
	app := x.NewApp()
	app.HandleJSON("/", func(c *x.Context) {
		c.Debug("디버그")
		c.Info("인포")
		c.Warn("워닝")
		c.Error("에러")
	})
	app.Run()
}
